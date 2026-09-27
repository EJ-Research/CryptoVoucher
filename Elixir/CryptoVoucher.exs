# Example usage: elixir CryptoVoucher.exs
Code.require_file("lib/crypto_voucher.ex", __DIR__)

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
