# Manage VM instance
resource "vstack_vm" "example" {
  name          = "example"
  description   = "This is a test VM for demonstration purposes."
  cpus          = 4
  ram           = 4096
  cpu_priority  = 10
  boot_media    = 0
  vcpu_class    = 1
  os_type       = 6
  os_profile    = "4001"
  vdc_id        = 1234
  pool_selector = "12345678911234567891"

  action = "start"

  disks = [
    {
      size       = 40
      slot       = 1
      label      = "Primary Disk"
      iops_limit = 0
      mbps_limit = 0
    }
  ]

  guest = {
    hostname          = "example"
    ssh_password_auth = 1
    users = {
      root = {
        ssh_authorized_keys = [
          "ssh-rsa AAAAB3NzaC1..."
        ]
        password = "password"
      }
    }
    resolver = {
      name_server = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
      search      = "example.local"
    }

    boot_cmds = ["swapoff -a"]

    run_cmds = ["systemctl restart ntpd"]
  }
}