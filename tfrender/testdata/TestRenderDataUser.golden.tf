terraform {
  required_providers {
    firehydrant = {
      source  = "firehydrant/firehydrant"
      version = ">= 0.7.1"
    }
  }
}

data "firehydrant_user" "user_0" {
  email = "user-0@example.com"
}

data "firehydrant_user" "user_1" {
  email = "user-1@example.com"
}

data "firehydrant_user" "user_2" {
  email = "user-2@example.com"
}
