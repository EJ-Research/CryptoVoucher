# Author: (EJ)
# Description: This code provides functionalities for encoding private keys into vouchers
# and restoring them using Base62 encoding. It is designed for flexibility and adaptability
# across multiple programming languages and platforms.
#
# License: MIT
# Feel free to use, modify, or distribute this code under the terms of the MIT License.

defmodule CryptoVoucher do
  @base62_chars "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
  # Byte -> Base62 index, built at compile time
  @base62_index @base62_chars |> :binary.bin_to_list() |> Enum.with_index() |> Map.new()
  # secp256k1 curve order; valid private keys are in [1, N-1]
  @secp256k1_n 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
  @voucher_key_length 28
  # 62^43 > 2^256, so every private key fits in 43 Base62 characters
  @encoded_length 43
  # Encoded key followed by one check character
  @voucher_length @encoded_length + 1

  # Validates if the input is a hexadecimal string of the specified length
  defp validate_hex(input, length) do
    if byte_size(input) != length do
      {:error, "Input must be #{length} characters long!"}
    else
      case Regex.match?(~r/\A[0-9a-fA-F]+\z/, input) do
        true -> :ok
        false -> {:error, "Input must be a valid hexadecimal string!"}
      end
    end
  end

  # Encodes a hexadecimal private key into a fixed-length Base62 string
  def base62_encode(hex_input) do
    case validate_hex(hex_input, 64) do
      :ok ->
        decimal_value = String.to_integer(hex_input, 16)

        if decimal_value > 0 and decimal_value < @secp256k1_n do
          base62 = encode_to_base62(decimal_value, "")
          # Leading zeros keep the length fixed and do not change the decoded value
          {:ok, String.pad_leading(base62, @encoded_length, "0")}
        else
          {:error, "Private key is out of secp256k1 range!"}
        end

      {:error, message} ->
        {:error, message}
    end
  end

  defp encode_to_base62(0, result), do: result
  defp encode_to_base62(decimal, result) do
    remainder = rem(decimal, 62)
    new_result = String.at(@base62_chars, remainder) <> result
    encode_to_base62(div(decimal, 62), new_result)
  end

  # Decodes a Base62 string back into a hexadecimal private key
  def base62_decode(base62_input) do
    cond do
      byte_size(base62_input) != @encoded_length ->
        {:error, "Base62 input must be #{@encoded_length} characters long!"}

      not Regex.match?(~r/\A[0-9A-Za-z]+\z/, base62_input) ->
        {:error, "Invalid Base62 input. Only alphanumeric characters are allowed!"}

      true ->
        decimal_value = decode_from_base62(base62_input, 0)

        if decimal_value > 0 and decimal_value < @secp256k1_n do
          {:ok, Base.encode16(<<decimal_value::256>>, case: :lower)}
        else
          {:error, "Private key is out of secp256k1 range!"}
        end
    end
  end

  # Input is already validated, so every byte is a Base62 character
  defp decode_from_base62("", result), do: result
  defp decode_from_base62(<<char, rest::binary>>, result) do
    decode_from_base62(rest, result * 62 + Map.fetch!(@base62_index, char))
  end

  # Computes the Luhn mod 62 check character of a validated Base62 string
  defp check_char(base62_input) do
    {total, _factor} =
      base62_input
      |> :binary.bin_to_list()
      |> Enum.reverse()
      |> Enum.reduce({0, 2}, fn char, {total, factor} ->
        addend = factor * Map.fetch!(@base62_index, char)
        {total + div(addend, 62) + rem(addend, 62), 3 - factor}
      end)

    binary_part(@base62_chars, rem(62 - rem(total, 62), 62), 1)
  end

  # Creates a voucher key and voucher code from a private key
  def create_voucher(private_key) do
    case base62_encode(private_key) do
      {:ok, base62_encoded} ->
        voucher = base62_encoded <> check_char(base62_encoded)
        voucher_key = binary_part(voucher, 0, @voucher_key_length)
        voucher_code = binary_part(voucher, @voucher_key_length, @voucher_length - @voucher_key_length)
        {:ok, voucher_key, voucher_code}

      {:error, message} ->
        {:error, message}
    end
  end

  # Restores a private key from a voucher key and voucher code
  def restore_private_key(voucher_key, voucher_code) do
    voucher = voucher_key <> voucher_code

    if byte_size(voucher) != @voucher_length do
      {:error, "Voucher must be #{@voucher_length} characters long!"}
    else
      <<base62_encoded::binary-size(@encoded_length), check::binary>> = voucher

      case base62_decode(base62_encoded) do
        {:ok, hex_output} ->
          if check_char(base62_encoded) == check do
            {:ok, hex_output}
          else
            {:error, "Invalid voucher check character!"}
          end

        {:error, message} ->
          {:error, message}
      end
    end
  end
end
