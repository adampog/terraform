provider "foo" {
  value = "bar"
}

module "subsub" {
    source = "./subsub"
}
