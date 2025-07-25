terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant-v2"
      version = "~> 0.3.0"
    }
  }
}

data "firehydrant_user" "user_0" {
  id = "id-for-user-0"
}

data "firehydrant_user" "user_1" {
  id = "id-for-user-1"
}

data "firehydrant_user" "user_2" {
  id = "id-for-user-2"
}
