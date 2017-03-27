package mocks

import pbr "tes/server/proto"

// SetupDefaultMockTemplates is a helper which sets up the mocked Client.Templates()
func (c *Client) SetupDefaultMockTemplates() {
	c.SetupMockTemplates(pbr.Resources{
		Cpus: 10.0,
		Ram:  100.0,
		Disk: 1000.0,
	})
}

// SetupMockTemplates sets the template returned by Client.Templates().
// The template will have the given resources.
func (c *Client) SetupMockTemplates(res pbr.Resources) {
	avail := res
	c.On("Templates").Return([]pbr.Worker{
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
	c.On("Templates").Return([]pbr.Worker{})
}
