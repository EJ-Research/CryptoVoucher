# Author: (EJ)
# Description: This code provides functionalities for encoding private keys into vouchers 
# and restoring them using Base62 encoding. It is designed for flexibility and adaptability 
# across multiple programming languages and platforms.
#
# License: MIT
# Feel free to use, modify, or distribute this code under the terms of the MIT License.


class CryptoVoucher
  BASE62_CHARS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

  # Validates if the input is a hexadecimal string of the specified length
  def validate_hex(input, length)
    if input.length != length
      return [false, "Input must be #{length} characters long!"]
    end
    unless input =~ /^[0-9a-fA-F]+$/
      return [false, "Input must be a valid hexadecimal string!"]
    end
    [true, nil]
  end

  # Encodes a hexadecimal string into a Base62 string
  def base62_encode(hex_input)
    valid, error = validate_hex(hex_input, 64)
    return [nil, error] unless valid

    decimal_value = hex_input.to_i(16)
    base62 = ""

    while decimal_value > 0
      remainder = decimal_value % 62
      base62 = BASE62_CHARS[remainder] + base62
      decimal_value /= 62
    end

    [base62, nil]
  end

  # Decodes a Base62 string back into a hexadecimal string
  def base62_decode(base62_input)
    unless base62_input =~ /^[0-9A-Za-z]+$/
      return [nil, "Invalid Base62 input. Only alphanumeric characters are allowed!"]
    end

    decimal_value = 0

    base62_input.each_char do |char|
      index = BASE62_CHARS.index(char)
      return [nil, "Invalid character in Base62 input."] if index.nil?

      decimal_value = decimal_value * 62 + index
    end

    hex_output = decimal_value.to_s(16).rjust(64, '0')
    [hex_output, nil]
  end

  # Creates a voucher key and voucher code from a private key
  def create_voucher(private_key)
    base62_encoded, error = base62_encode(private_key)
    return [nil, nil, error] if error

    if base62_encoded.length < 28
      return [nil, nil, "Encoded key is too short!"]
    end

    voucher_key = base62_encoded[0, 28]
    voucher_code = base62_encoded[28..]
    [voucher_key, voucher_code, nil]
  end

  # Restores a private key from a voucher key and voucher code
  def restore_private_key(voucher_key, voucher_code)
    combined = voucher_key + voucher_code
    base62_decode(combined)
  end
end

# Example usage
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