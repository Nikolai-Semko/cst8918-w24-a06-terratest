# Define output values for later reference
output "resource_group_name" {
  value = azurerm_resource_group.rg.name
}

output "vm_name" {
  value = azurerm_linux_virtual_machine.webserver.name
}

output "nic_name" {
  value = azurerm_network_interface.webserver.name
}

output "public_ip" {
  value = azurerm_public_ip.webserver.ip_address  # CHANGED - get IP from public_ip resource
}