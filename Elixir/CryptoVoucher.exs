# Author: (EJ)
# Description: This code provides functionalities for encoding private keys into vouchers 
# and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
# across multiple programming languages and platforms.
#
# License: MIT
# Feel free to use, modify, or distribute this code under the terms of the MIT License.

defmodule CryptoVoucher do
  @base62_chars "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

  # Validates if the input is a hexadecimal string of the specified length
  defp validate_hex(input, length) do
    if String.length(input) != length do
      {:error, "Input must be #{length} characters long!"}
    else
      case Regex.match?(~r/^[0-9a-fA-F]+$/, input) do
        true -> :ok
        false -> {:error, "Input must be a valid hexadecimal string!"}
      end
    end
  end

  # Encodes a hexadecimal string into a Base62 string
  def base62_encode(hex_input) do
    case validate_hex(hex_input, 64) do
      :ok ->
        decimal_value = String.to_integer(hex_input, 16)
        base62 = encode_to_base62(decimal_value, "")
        {:ok, base62}

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

  # Decodes a Base62 string back into a hexadecimal string
  def base62_decode(base62_input) do
    if Regex.match?(~r/^[0-9A-Za-z]+$/, base62_input) do
      decimal_value = decode_from_base62(base62_input, 0)
      hex_output = Integer.to_string(decimal_value, 16) |> String.pad_leading(64, "0")
      {:ok, hex_output}
    else
      {:error, "Invalid Base62 input. Only alphanumeric characters are allowed!"}
    end
  end

  defp decode_from_base62("", result), do: result
  defp decode_from_base62(<<char, rest::binary>>, result) do
    index = Enum.find_index(@base62_chars |> String.graphemes(), fn x -> x == <<char>> end)

    if index == nil do
      {:error, "Invalid character in Base62 input!"}
    else
      decode_from_base62(rest, result * 62 + index)
    end
  end

  # Creates a voucher key and voucher code from a private key
  def create_voucher(private_key) do
    case base62_encode(private_key) do
      {:ok, base62_encoded} ->
        if String.length(base62_encoded) < 28 do
          {:error, "Encoded key is too short."}
        else
          voucher_key = String.slice(base62_encoded, 0..27)
          voucher_code = String.slice(base62_encoded, 28..-1)
          {:ok, voucher_key, voucher_code}
        end

      {:error, message} ->
        {:error, message}
    end
  end

  # Restores a private key from a voucher key and voucher code
  def restore_private_key(voucher_key, voucher_code) do
    combined = voucher_key <> voucher_code

    case base62_decode(combined) do
      {:ok, hex_output} -> {:ok, hex_output}
      {:error, message} -> {:error, message}
    end
  end
end

# Example usage
private_key = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B"

case CryptoVoucher.create_voucher(private_key) do
  {:ok, voucher_key, voucher_code} ->
    IO.puts("Voucher Key: #{voucher_key}")
    IO.puts("Voucher Code: #{voucher_code}")

    case CryptoVoucher.restore_private_key(voucher_key, voucher_code) do
      {:ok, restored_key} ->
        IO.puts("Restored Private Key: #{String.upcase(restored_key)}")
        if String.upcase(restored_key) == String.upcase(private_key) do
          IO.puts("Success!")
        else
          IO.puts("Failed!")
        end

      {:error, message} ->
        IO.puts("Error restoring private key: #{message}")
    end

  {:error, message} ->
    IO.puts("Error creating voucher: #{message}")
end
