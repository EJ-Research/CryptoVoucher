# Author: (EJ)
# Description: This code provides functionalities for encoding private keys into vouchers 
# and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
# across multiple programming languages and platforms.
#
# License: MIT
# Feel free to use, modify, or distribute this code under the terms of the MIT License.
#
# DEPRECATED: The Ruby implementation is no longer maintained. It produces the same
# vouchers as the Go, Node.js, Python and PHP versions, but will not receive updates.

warn "[DEPRECATION] CryptoVoucher for Ruby is no longer maintained. Use the Go, Node.js, Python or PHP version."

class CryptoVoucher
  BASE62_CHARS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
  # secp256k1 curve order; valid private keys are in [1, N-1]
  SECP256K1_N = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
  VOUCHER_KEY_LENGTH = 28
  # 62^43 > 2^256, so every private key fits in 43 Base62 characters
  ENCODED_LENGTH = 43
  # Remaining Base62 characters followed by one check character
  VOUCHER_CODE_LENGTH = ENCODED_LENGTH - VOUCHER_KEY_LENGTH + 1

  # Validates if the input is a hexadecimal string of the specified length
  def validate_hex(input, length)
    if input.length != length
      return [false, "Input must be #{length} characters long!"]
    end
    unless input =~ /\A[0-9a-fA-F]+\z/
      return [false, "Input must be a valid hexadecimal string!"]
    end
    [true, nil]
  end

  # Encodes a hexadecimal private key into a fixed-length Base62 string
  def base62_encode(hex_input)
    valid, error = validate_hex(hex_input, 64)
    return [nil, error] unless valid

    decimal_value = hex_input.to_i(16)
    unless decimal_value > 0 && decimal_value < SECP256K1_N
      return [nil, "Private key is out of secp256k1 range!"]
    end

    base62 = ""

    while decimal_value > 0
      remainder = decimal_value % 62
      base62 = BASE62_CHARS[remainder] + base62
      decimal_value /= 62
    end

    # Leading zeros keep the length fixed and do not change the decoded value
    [base62.rjust(ENCODED_LENGTH, '0'), nil]
  end

  # Decodes a Base62 string back into a hexadecimal private key
  def base62_decode(base62_input)
    if base62_input.length != ENCODED_LENGTH
      return [nil, "Base62 input must be #{ENCODED_LENGTH} characters long!"]
    end
    unless base62_input =~ /\A[0-9A-Za-z]+\z/
      return [nil, "Invalid Base62 input. Only alphanumeric characters are allowed!"]
    end

    decimal_value = 0

    base62_input.each_char do |char|
      index = BASE62_CHARS.index(char)
      return [nil, "Invalid character in Base62 input."] if index.nil?

      decimal_value = decimal_value * 62 + index
    end

    unless decimal_value > 0 && decimal_value < SECP256K1_N
      return [nil, "Private key is out of secp256k1 range!"]
    end

    hex_output = decimal_value.to_s(16).rjust(64, '0')
    [hex_output, nil]
  end

  # Computes the check character of a validated Base62 string: the sum of each digit
  # times its 1-based position, modulo 61
  def check_char(base62_input)
    total = 0
    base62_input.each_char.with_index(1) do |char, position|
      total += position * BASE62_CHARS.index(char)
    end
    BASE62_CHARS[total % 61]
  end

  # Creates a voucher key and voucher code from a private key
  def create_voucher(private_key)
    base62_encoded, error = base62_encode(private_key)
    return [nil, nil, error] if error

    voucher = base62_encoded + check_char(base62_encoded)
    voucher_key = voucher[0, VOUCHER_KEY_LENGTH]
    voucher_code = voucher[VOUCHER_KEY_LENGTH..]
    [voucher_key, voucher_code, nil]
  end

  # Restores a private key from a voucher key and voucher code
  def restore_private_key(voucher_key, voucher_code)
    if voucher_key.length != VOUCHER_KEY_LENGTH
      return [nil, "Voucher key must be #{VOUCHER_KEY_LENGTH} characters long!"]
    end
    if voucher_code.length != VOUCHER_CODE_LENGTH
      return [nil, "Voucher code must be #{VOUCHER_CODE_LENGTH} characters long!"]
    end

    voucher = voucher_key + voucher_code

    base62_encoded = voucher[0, ENCODED_LENGTH]
    hex_output, error = base62_decode(base62_encoded)
    return [nil, error] if error

    if check_char(base62_encoded) != voucher[ENCODED_LENGTH]
      return [nil, "Invalid voucher check character!"]
    end

    [hex_output, nil]
  end
end

# Example usage (runs only when this file is executed directly)
if __FILE__ == $PROGRAM_NAME
  private_key = "0B6BF630452AABF9C57A2755DD4B3DD570A4047181C8A3A44239AD50E9F7D06B"
  crypto_voucher = CryptoVoucher.new

  # Create a voucher
  voucher_key, voucher_code, error = crypto_voucher.create_voucher(private_key)
  if error
    puts "Error creating voucher: #{error}"
  else
    puts "Voucher Key: #{voucher_key}"
    puts "Voucher Code: #{voucher_code}"

    # Restore the private key
    restored_key, error = crypto_voucher.restore_private_key(voucher_key, voucher_code)
    if error
      puts "Error restoring private key: #{error}"
    else
      puts "Restored Private Key: #{restored_key.upcase}"
      if restored_key.upcase == private_key.upcase
        puts "Success!"
      else
        puts "Failed!"
      end
    end
  end
end
