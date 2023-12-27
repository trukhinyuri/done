provider "kubernetes" {
  # Конфигурация провайдера (например, endpoint, token, и т.д.)
}

variable "pods" {
  description = "Список конфигураций для создания подов"
  type = list(object({
    name      = string
    image     = string
    namespace = string
  }))
  default = [
    {
      name      = "pod1"
      image     = "nginx"
      namespace = "default"
    },
    {
      name      = "pod2"
      image     = "busybox"
      namespace = "default"
    }
    # Добавьте больше подов здесь, если нужно
  ]
}

resource "kubernetes_pod" "example" {
  for_each = { for pod in var.pods : pod.name => pod }

  metadata {
    name      = each.value.name
    namespace = each.value.namespace
  }

  spec {
    container {
      image = each.value.image

      name  = each.value.name
    }
  }
}
