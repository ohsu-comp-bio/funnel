package mocks

import pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"

// SetupDefaultMockTemplates is a helper which sets up the mocked Client.Templates()
func (c *Client) SetupDefaultMockTemplates() {
	c.SetupMockTemplates(pbf.Resources{
		Cpus:   10.0,
		RamGb:  100.0,
		DiskGb: 1000.0,
	})
}

// SetupMockTemplates sets the template returned by Client.Templates().
// The template will have the given resources.
func (c *Client) SetupMockTemplates(res pbf.Resources) {
	avail := res
	c.On("Templates").Return([]pbf.Worker{
		{
			Metadata: map[string]string{
				"gce":          "yes",
				"gce-template": "test-tpl",
			},
			Resources: &res,
			Available: &avail,
		},
	})
}

// SetupEmptyMockTemplates sets the mock to return an empty slice from
// Client.Templates()
func (c *Client) SetupEmptyMockTemplates() {
	c.On("Templates").Return([]pbf.Worker{})
}
