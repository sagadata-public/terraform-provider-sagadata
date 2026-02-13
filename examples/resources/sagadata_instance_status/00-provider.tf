terraform {
  required_providers {
    sagadata = {
      source = "sagadata-public/sagadata"
    }
  }
}

provider "sagadata" {
  # optional configuration...
}
