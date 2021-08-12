provider "vsphere" {
  user           = "user"
  password       = "pass"
  vsphere_server = "10.10.10.10"

  allow_unverified_ssl = false
}
