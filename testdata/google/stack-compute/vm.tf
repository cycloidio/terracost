module "vm" {
  #####################################
  # Do not modify the following lines #
  source = "./module-gce"

  project  = var.project
  env      = var.env
  customer = var.customer

  ##################################### 

  ###
  # FIREWALL 
  ###

  ## Firewall - ingress

  #. ingress_firewall_name (optional, string):
  #+ Name of the firewall ingress resource.
  ingress_firewall_name = "${var.customer}-${var.project}-${var.env}-ingress"

  #. ingress_disabled (optional, bool):
  #+ Denotes whether the firewall ingress rule is disabled.
  ingress_disabled = false

  #. ingress_source_ranges (optional, array):
  #+ If specified the firewall will only be applied to the source IP address in these ranges.
  ingress_source_ranges = []

  #. ingress_source_tags (optional, array):
  #+ If source tags are specified, the firewall will apply only to traffic with source IP that belongs to a tag listed in source tags. 
  ingress_source_tags = []

  #. ingress_allow_protocol (optional, string):
  #+ The IP protocol to which the ingress allow rule applies.
  ingress_allow_protocol = "tcp"
  
  #. ingress_allow_ports (optional, map):
  #+ An otional list of IP ports to which the ingress allow rule applies.
  ingress_allow_ports = ["22"]

  ## Firewall - egress

  #. egress_firewall_name (optional, string):
  #+ Name of the firewall egress resource.
  egress_firewall_name = "${var.customer}-${var.project}-${var.env}-egress"

  #. egress_disabled (optional, bool):
  #+ Denotes whether the firewall egress rule is disabled.
  egress_disabled = true

  #. egress_destination_ranges (optional, array):
  #+ Lists the destination IP address, as CIDR, to apply egress firewalls rules.
  egress_destination_ranges = []

  #. egress_allow_protocol (optional, string):
  #+ The IP protocol to which the egress allow rule applies.
  egress_allow_protocol = ""

  #. ingress_allow_ports (optional, map):
  #+ An otional list of IP ports to which the egress allow rule applies.
  egress_allow_ports = [""]


  ###
  # Cloud init template
  ###
  #. file_content (optional, string):
  #+ The content of the file to use if cloud init is used.
  file_content = ""


  ###
  # Virtual Machine
  ###

  # VM - general speficiations

  #. instance_name (required, string):
  #+ The unique name for the instance.
  instance_name = "${var.customer}-${var.project}-${var.env}-vm"

  #. machine_type (required, string):
  #+ The machine type to create.
  machine_type = "e2-small"

  #. allow_stopping_for_update (optional, bool):
  #+ Allows to stop the instance to update its properties.
  allow_stopping_for_update = true

  # VM - tags

  #. instance_tags (optional, array):
  #+ A list of network tags to attach to the instance.
  instance_tags = ["${var.customer}-${var.project}-${var.env}-network-tag"]


  # VM - storage

  #. boot_disk_auto_delete (optional, bool):
  #+ Enables disk deletion when the instance is deleted.
  boot_disk_auto_delete = true

  #. boot_disk_device_name (optional, string):
  #+ Name with which attached disk will be accessible, as /dev/disk/by-id/google-{{device_name}}.
  boot_disk_device_name = ""

  #. boot_disk_image (optional, string):
  #+ The image from which to initialize this disk.
  boot_disk_image = "debian-cloud/debian-10"

  #. boot_disk_size (optional, integer):
  #+ The size of the image in gigabytes. If not specified, it will inherit the size of its base image.
  boot_disk_size = 5

  #. boot_disk_type (optional, string):
  #+ The GCE disk type
  boot_disk_type = ""


  # VM - network

  #. network (required, string):
  #+ The name or self_link of the network to attach this interface to.
  network = "network_name_to_specify"

  #. network_ip (optional, string):
  #+ The private IP address to assign to the instance. If empty, the address will be automatically assigned.
  network_ip = ""


  # VM - labels

  #. instance_extra_labels (optional):
  #+ A map of key/value label pairs to assign to the instance.
  instance_extra_labels = {}
}
