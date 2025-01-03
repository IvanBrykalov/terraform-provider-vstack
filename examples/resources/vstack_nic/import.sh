# NIC instance can be imported by specifying the identifier with following format "vm_id/id"
# vm_id - is the ID of the virtual machine that owns the instance of NIC that you want to import.
# id - is ID of NIC that you want to import

terraform import vstack_nic.example_nic1 1234/5678