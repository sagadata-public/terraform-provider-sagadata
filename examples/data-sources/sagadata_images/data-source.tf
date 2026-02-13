data "sagadata_images" "cloud-images" {
  filter = {
    type = "cloud-image"
  }
}

data "sagadata_images" "snapshots" {
  filter = {
    type   = "snapshot"
    region = "NORD-NO-KRS-1"
  }
}

data "sagadata_images" "preconfigured-images" {
  filter = {
    type = "preconfigured"
  }
}
