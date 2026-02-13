resource "sagadata_instance" "example" {
  name   = "example"
  region = "NORD-NO-KRS-1"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}


resource "sagadata_instance_status" "example" {
  instance_id = sagadata_instance.example.id
  status      = "active"
}
