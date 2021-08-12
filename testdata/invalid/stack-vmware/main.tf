provider "vsphere" {
  user           = "user"
  password       = "pass"
  vsphere_server = "10.10.10.10"

  allow_unverified_ssl = false
}

module "example-server-linuxvm-withdatadisk" {
  source             = "Terraform-VMWare-Modules/vm3nic/vsphere"
  version            = "0.1.0"
  vmtemp             = "TemplateName"
  instances          = 1
  vmname             = "example-server-windows"
  vmrp               = "esxi/Resources"
  net01              = "Name of the VLAN in vSphere for the first NIC"
  net02              = "Name of the VLAN in vSphere for the Second NIC"
  net03              = "Name of the VLAN in vSphere for the Third NIC"
  data_disk          = "true"
  data_disk_size_gb  = 20
  dc                 = "Datacenter"
  ds                 = "Data Store name"
}
