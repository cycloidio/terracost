###
# FIREWALL 
###

resource "google_compute_firewall" "ingress" {
  name          = local.ingress_name
  network       = var.network
  description   = var.ingress_firewall_description
  direction     = "INGRESS"
  disabled      = var.ingress_disabled
  source_ranges = var.ingress_source_ranges
  source_tags   = var.ingress_source_tags
  target_tags   = local.instance_tags

  allow {
    protocol = var.ingress_allow_protocol
    ports    = var.ingress_allow_ports
  }
}

resource "google_compute_firewall" "egress" {
  name               = local.egress_name
  network            = var.network
  direction          = "EGRESS"
  disabled           = var.egress_disabled
  destination_ranges = var.egress_destination_ranges
  target_tags        = local.instance_tags

  allow {
    protocol = var.egress_allow_protocol
    ports    = var.egress_allow_ports
  }
}

###
# Cloud init template
###
data "template_file" "user_data" {
  template = file("${path.module}/cloud-init.sh.tpl")
  vars = {
    file_content  = var.file_content
  }
}

###
# Virtual Machine
###

resource "google_compute_instance" "vm" {
  # general configuration
  name                      = local.instance_name
  description               = var.instance_description
  machine_type              = var.machine_type
  allow_stopping_for_update = var.allow_stopping_for_update

  # tags
  tags = local.instance_tags

  # storage
  boot_disk {
    auto_delete = var.boot_disk_auto_delete
    device_name = var.boot_disk_device_name
    initialize_params {
      image = var.boot_disk_image
      size  = var.boot_disk_size
      type  = var.boot_disk_type
    }
  }

  # cloud-init
  metadata_startup_script = data.template_file.user_data.rendered

  # network
  network_interface {
    network    = var.network
    network_ip = var.network_ip

    // If omited the instance is not accessible from the Internet. Provides an ephemeral public IP
    access_config {

    }
  }
  // Required in case user attaches other disk
  // When using google_compute_attached_disk you must use lifecycle.ignore_changes = ["attached_disk"] 
  // on the google_compute_instance resource that has the disks attached. 
  // Otherwise the two resources will fight for control of the attached disk block.
  lifecycle {
    ignore_changes = [attached_disk]
  }

  # labels
  labels = local.instance_labels
}