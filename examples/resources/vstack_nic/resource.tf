# Manage VM NIC
resource "vstack_nic" "example_nic1" {
  vm_id           = vstack_vm.example_vm.id
  network_id      = 1234
  slot            = 1
  address         = "192.168.0.2"
  ratelimit_mbits = 0
  depends_on      = [vstack_vm.example_vm]
}