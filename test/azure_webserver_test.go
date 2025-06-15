package test

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "100148d9-6e2a-4efd-b476-267627a09dfa"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "semk0001",
		},
		// Use terraform binary instead of tofu
		TerraformBinary: "terraform",
	}

	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of output variables
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name")

	// Original test: Confirm VM exists
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// Test 1: Confirm NIC exists and is connected to VM
	t.Run("Test_NIC_Exists_And_Connected", func(t *testing.T) {
		// Check if NIC exists
		nicExists := azure.NetworkInterfaceExists(t, nicName, resourceGroupName, subscriptionID)
		assert.True(t, nicExists, "Network Interface should exist")

		// Get VM details to check NIC attachment
		vm := azure.GetVirtualMachine(t, vmName, resourceGroupName, subscriptionID)
		require.NotNil(t, vm, "VM should not be nil")
		require.NotNil(t, vm.NetworkProfile, "VM NetworkProfile should not be nil")
		require.NotEmpty(t, vm.NetworkProfile.NetworkInterfaces, "VM should have network interfaces")

		// Check if our NIC is attached to the VM
		nicFound := false
		for _, nic := range *vm.NetworkProfile.NetworkInterfaces {
			if nic.ID != nil && strings.Contains(*nic.ID, nicName) {
				nicFound = true
				t.Logf("Found NIC %s attached to VM %s", nicName, vmName)
				break
			}
		}
		assert.True(t, nicFound, "NIC should be attached to VM")
	})

	// Test 2: Confirm the VM is running the correct Ubuntu version
	t.Run("Test_Ubuntu_Version", func(t *testing.T) {
		// Get VM details
		vm := azure.GetVirtualMachine(t, vmName, resourceGroupName, subscriptionID)
		require.NotNil(t, vm, "VM should not be nil")
		require.NotNil(t, vm.StorageProfile, "VM StorageProfile should not be nil")
		require.NotNil(t, vm.StorageProfile.ImageReference, "VM ImageReference should not be nil")

		imageRef := vm.StorageProfile.ImageReference

		// Check Publisher
		require.NotNil(t, imageRef.Publisher, "Publisher should not be nil")
		assert.Equal(t, "Canonical", *imageRef.Publisher, "Publisher should be Canonical")

		// Check Offer (Ubuntu Server)
		require.NotNil(t, imageRef.Offer, "Offer should not be nil")
		assert.Equal(t, "0001-com-ubuntu-server-jammy", *imageRef.Offer, "Should be Ubuntu Server Jammy")

		// Check SKU (22.04 LTS)
		require.NotNil(t, imageRef.Sku, "SKU should not be nil")
		assert.Equal(t, "22_04-lts-gen2", *imageRef.Sku, "Should be Ubuntu 22.04 LTS Gen2")

		// Log the version info
		t.Logf("VM is running: Publisher=%s, Offer=%s, SKU=%s",
			*imageRef.Publisher, *imageRef.Offer, *imageRef.Sku)
	})
}
