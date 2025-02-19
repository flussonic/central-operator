
# central-operator

// TODO(user): Add simple overview of use/purpose

## Description

// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

**Действия для запуска кластера из нескольких инстансов централа и флюссоником.**

Примените операторы централа и медиасервера, чтобы k8s имел представление о CRD и в нем запустились экземпляры операторов, ждущие появления в кластере соответствующих ресурсов:

```sh
kubectl apply -f https://flussonic.github.io/media-server-operator/latest/operator.yaml
kubectl apply -f https://flussonic.github.io/central-operator/latest/operator.yaml
```

После этого нужно добавить на ноды соответствующие лейблы и добавить необходимые секреты:

```sh
kubectl label nodes node_name flussonic.com/streamer=true
kubectl create secret generic flussonic-license \
    --from-literal=license_key="${LICENSE_KEY}" \
    --from-literal=edit_auth="root:password"
```

\* node_name - имя ноды, на которую мы хотим, чтобы заехал флюссоник. Копия флюссоника заедет на каждую ноду, промаркированную соответствующим лейблом.

После чего необходимо добавить нужные кастомные ресурсы, чтобы операторы развернули и начали провижнить соответствующие стандартные ресурсы kubernetes:

```sh
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/ingress.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/media_valpha1_mediaserver.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/media_valpha1_central.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/postgres.yaml
kubectl apply -f https://raw.githubusercontent.com/flussonic/central-operator/master/config/samples/redis.yaml
```

Примечание: сейчас централ корректно работает только с nginx в качестве ingess-контроллера (требование агента), поэтому если по умолчанию используется что-то другое, необходимо удалить его из кубера и применить манифесты ingress-nginx:

```sh
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
```

При выполнении всех инструкций выше должен был развернуться кластер из постгреса, редиса, двух инстансов централа и одного медиасервера.

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Для разработчиков

Проект сгенерирован при помощи [operator-sdk](https://sdk.operatorframework.io/).

Чтобы сгенерить доку нужно выполнить в makefile исправить `version` на актуальную, и выполнить `make operator.yaml`.

Для запуска кластера на multipass и k3s необходимо выполнить `make mp-start`. В скрипт зашито то, что контроллер билдится локально и прокидывается на ноды.

Если необходимо запустить кластер с образом контроллера из докерхаба, можно выполнить `./mp-start.sh`.
