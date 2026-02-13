resource "sagadata_instance" "target" {
  # ...
}

resource "sagadata_snapshot" "example" {
  name        = "example"
  instance_id = sagadata_instance.target.id

  retain_on_delete = true # optional
}
