package mocks

import "google.golang.org/api/compute/v1"

// SetupMockMachineTypes is a helper to set up the wrapper mock to return
// the a common machine type list from Wrapper.ListMachineTypes()
func (wpr *Wrapper) SetupMockMachineTypes() {
	wpr.On("ListMachineTypes", "test-proj", "test-zone").Return(&compute.MachineTypeList{
		Items: []*compute.MachineType{
			{
				Name:      "test-mt",
				GuestCpus: 3,
				MemoryMb:  12,
			},
		},
	}, nil)
}

// SetupMockInstanceTemplates is a helper which helps set up the wrapper mock
// to return a common instance template list.
func (wpr *Wrapper) SetupMockInstanceTemplates() {
	wpr.On("ListInstanceTemplates", "test-proj").Return(&compute.InstanceTemplateList{
		Items: []*compute.InstanceTemplate{
			{
				Name: "test-tpl",
				Properties: &compute.InstanceProperties{
					MachineType: "test-mt",
					Disks: []*compute.AttachedDisk{
						{
							InitializeParams: &compute.AttachedDiskInitializeParams{
								DiskSizeGb: 14,
							},
						},
					},
					Metadata: &compute.Metadata{},
					Tags: &compute.Tags{
						Items: []string{"funnel"},
					},
				},
			},
		},
	}, nil)
}
