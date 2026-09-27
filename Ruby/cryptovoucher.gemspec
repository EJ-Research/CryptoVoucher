Gem::Specification.new do |spec|
  spec.name = "cryptovoucher"
  spec.version = "1.0.0"
  spec.summary = "Deprecated: encode private keys into Base62 vouchers and restore them"
  spec.post_install_message = "CryptoVoucher for Ruby is no longer maintained. Use the Go, Node.js, Python or PHP version."
  spec.authors = ["EJ"]
  spec.license = "MIT"
  spec.homepage = "https://github.com/EJ-Research/CryptoVoucher"
  spec.required_ruby_version = ">= 2.7"
  spec.files = ["CryptoVoucher.rb", "lib/cryptovoucher.rb"]
  spec.require_paths = ["lib", "."]
end
