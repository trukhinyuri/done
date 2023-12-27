provider "kubernetes" {
  config_path = "~/.kube/config"
}

import {
  //Resource_Address
  to = kubernetes_pod.App

  //Unique_Identifier
  id = "default/example-pod"
}




resource "kubernetes_pod" "App" {
  metadata {
    name   = "example-pod"
    labels = {
      App = "example-app"
    }
  }

  spec {
    container {
      image = "nginx:1.7.8"
      name  = "nginx"
    }
  }
}