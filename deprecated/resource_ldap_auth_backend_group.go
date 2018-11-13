package deprecated

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/hashicorp/vault/api"
)

func ldapAuthBackendGroupResource() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		Create: ldapAuthBackendGroupResourceWrite,
		Update: ldapAuthBackendGroupResourceWrite,
		Read:   ldapAuthBackendGroupResourceRead,
		Delete: ldapAuthBackendGroupResourceDelete,
		Exists: ldapAuthBackendGroupResourceExists,

		Schema: map[string]*schema.Schema{
			"groupname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policies": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},
			"backend": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "ldap",
				StateFunc: func(v interface{}) string {
					return strings.Trim(v.(string), "/")
				},
			},
		},
	}
}

func ldapAuthBackendGroupResourcePath(backend, groupname string) string {
	return "auth/" + strings.Trim(backend, "/") + "/groups/" + strings.Trim(groupname, "/")
}

func ldapAuthBackendGroupResourceWrite(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	backend := d.Get("backend").(string)
	groupname := d.Get("groupname").(string)
	path := ldapAuthBackendGroupResourcePath(backend, groupname)

	data := map[string]interface{}{}

	if v, ok := d.GetOk("policies"); ok {
		data["policies"] = v.(*schema.Set).List()
	}

	log.Printf("[DEBUG] Writing LDAP group %q", path)
	_, err := client.Logical().Write(path, data)

	d.SetId(path)

	if err != nil {
		d.SetId("")
		return fmt.Errorf("error writing ldap group %q: %s", path, err)
	}
	log.Printf("[DEBUG] Wrote LDAP group %q", path)

	return ldapAuthBackendGroupResourceRead(d, meta)
}

func ldapAuthBackendGroupResourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	path := d.Id()

	log.Printf("[DEBUG] Reading LDAP group %q", path)
	resp, err := client.Logical().Read(path)
	if err != nil {
		return fmt.Errorf("error reading ldap group %q: %s", path, err)
	}
	log.Printf("[DEBUG] Read LDAP group %q", path)

	if resp == nil {
		log.Printf("[WARN] LDAP group %q not found, removing from state", path)
		d.SetId("")
		return nil
	}

	d.Set("policies",
		schema.NewSet(
			schema.HashString, resp.Data["policies"].([]interface{})))

	return nil

}

func ldapAuthBackendGroupResourceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	path := d.Id()

	log.Printf("[DEBUG] Deleting LDAP group %q", path)
	_, err := client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("error deleting ldap group %q", path)
	}
	log.Printf("[DEBUG] Deleted LDAP group %q", path)

	return nil
}

func ldapAuthBackendGroupResourceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*api.Client)
	path := d.Id()

	log.Printf("[DEBUG] Checking if LDAP group %q exists", path)
	resp, err := client.Logical().Read(path)
	if err != nil {
		return true, fmt.Errorf("error checking for existence of ldap group %q: %s", path, err)
	}
	log.Printf("[DEBUG] Checked if LDAP group %q exists", path)

	return resp != nil, nil
}