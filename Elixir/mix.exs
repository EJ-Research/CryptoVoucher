defmodule CryptoVoucher.MixProject do
  use Mix.Project

  def project do
    [
      app: :crypto_voucher,
      version: "1.0.0",
      elixir: "~> 1.12",
      deps: [],
      description: "Encode private keys into Base62 vouchers and restore them",
      package: [
        licenses: ["MIT"],
        links: %{"GitHub" => "https://github.com/EJ-Research/CryptoVoucher"}
      ]
    ]
  end
end
