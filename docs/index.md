---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "vstack Provider"
subcategory: ""
description: |-
  
---

# vstack Provider



## Example Usage

```terraform
provider "vstack" {
  username = "user"
  password = "password"
  host     = "https://vstack.example.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `host` (String) API host for the Vstack provider.
- `password` (String, Sensitive) Password for Vstack API.
- `username` (String) Username for Vstack API.