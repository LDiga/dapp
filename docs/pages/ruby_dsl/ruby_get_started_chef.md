---
title: Первое приложение на dapp
sidebar: ruby_dsl
permalink: ruby
---

В этой главе описана сборка простейшего приложения с помощью dapp. Перед изучением dapp желательно представлять, что такое Dockerfile и его основные [директивы](https://docs.docker.io/).

Для запуска примеров понадобятся:

* dapp (установка описана [здесь](./installation.html))
* docker версии не ниже 1.10
* git

## Сборка простого приложения

Начнём с простого приложения на php. Создайте директорию для тестов и склонируйте репозиторий:

```
git clone https://github.com/awslabs/opsworks-demo-php-simple-app
```

Это совсем небольшое приложение с одной страницей и статическими файлами. Чтобы приложение можно было запустить, нужно запаковать его в контейнер, например, с php и apache. Для этого достаточно такого Dockerfile.

```
$ vi Dockerfile
FROM php:7.0-apache

COPY . /var/www/html/

EXPOSE 80
EXPOSE 443
```

Соберите и запустите приложение:

```
$ docker build -t simple-app-v1 .
$ docker run -d --rm --name simple-app simple-app-v1
```

Проверить как работает приложение можно либо зайдя браузером на порт 80, либо выполнив curl внутри контейнера:

```
$ docker exec -ti simple-app bash
root@da234e2a7777:/var/www/html# curl 127.0.0.1
...
                <h1>Simple PHP App</h1>
                <h2>Congratulations!</h2>
...
```

Остановите контейнер с приложением:

```
docker stop simple-app
```

## Сборка с dapp

Теперь соберём образ приложения с помощью dapp. Для этого нужно создать dappfile.yaml - который будет содержать команды для сборки образа используя YAML синтаксис.

* В репозитории могут находится одновременно и dappfile.yaml и Dockerfile - они друг другу не мешают.
* Среди директив dappfile.yaml есть семейство docker.* директив, которые повторяют аналогичные из Dockerfile.

Создайте в корневой папке кода приложения фаил `dappfile.yaml` следующего содержания:
```
dimg: simple-php-app
from: php:7.0-apache
docker:
  EXPOSE: '80'
  EXPOSE: '443'
git:
- add: '/'
  to: '/var/www/html'
  includePaths:
  - '*.php'
  - 'assets'
```

Рассмотрим подробнее этот файл.

`dimg` — эта директива определяет тип и название образа, который будет собран (вместо `dimg` может быть `artifact` для типа образа - артефакт). Аргумент `simple-php-app` — имя этого образа, его можно увидеть, запустив `dapp dimg list`. Блок с остальными директивами определяет шаги для сборки образа.

`from` — аналог директивы `FROM`. Определяет базовый образ, на основе которого будет собираться образ приложения.

 [Подробнее](directives_images.html) про директивы `dimg` и `from`.

`git` — директива, на первый взгляд аналог директив `ADD` или `COPY`, но с более тесной интеграцией с git. Подробнее про то, как dapp работает с git, можно [прочесть](git.html) в отдельной главе, а сейчас главное увидеть, что директива `git` и вложенная директива `add` позволяют копировать содержимое локального git-репозитория в образ. Копирование производится из пути, указанного в `add`. `'/'` означает, что копировать нужно из корня репозитория. `to` задаёт конечную директорию в образе, куда попадут файлы. С помощью `includePaths` и `excludePaths` можно задавать, какие именно файлы нужно скопировать или какие нужно пропустить.

Для сборки выполните команду `dapp dimg build`

```
$ dapp dimg build
simple-php-app
  simple-php-app: calculating stages signatures                                                                      [RUNNING]
    Repository `own`: latest commit `c24fbabde24014459c907f2e734f701d4506eb08` to `/var/www/html`
  simple-php-app: calculating stages signatures                                                                           [OK] 0.17 sec
  From ...                                                                                                                [OK] 0.5 sec
    signature: dimgstage-opsworks-demo-php-simple-app:7967f820a9118bd7f453ff9ecd611678fcde38afc64be832a021548db6540ed7
  Git artifacts: create archive ...                                                                                       [OK] 0.4 sec
    signature: dimgstage-opsworks-demo-php-simple-app:3c0630c7f7b16e7bb99c186af48f5f010d5cf4fc4e8940d003fdc5a6237eb271
  Setup group
    Git artifacts: apply patches (after setup) ...                                                                        [OK] 0.39 sec
      signature: dimgstage-opsworks-demo-php-simple-app:fbda7c00c141bb3991d4695987383d55730b9d52e2e5c4f661c5f5c7d8692421
  Docker instructions ...                                                                                                 [OK] 0.37 sec
    signature: dimgstage-opsworks-demo-php-simple-app:7a6dc3b72c3dd76ce2b4201bc4cab86f4907d964f1bab6b1c206a710ac835b25
    instructions:
      EXPOSE 443
Running time 2.85 seconds
```

Запустите контейнер приложения из собранного образа с помощью команды `dapp dimg run`.

```
$ dapp dimg run -d --rm
59ae767d497b4e4fb8c32cd97110cc0f17e67d8e3c7f540cef73b713ef995e5a
```

Ключи, передаваемые в конце команды `dapp dimg run` (в примере это ключи `-d` и `--rm`) передаются в docker без изменений ([подробнее](dimg_run.html)).

Теперь можно проверить работу приложения как и ранее, только указывать контейнер нужно по его ID, а не имени (т.к. dapp не присваивал имя созданному контейнеру приложения). Узнать ID запущенного контейнера можно с помощью команды `docker ps` - ID контейнера будет в колонке с заголовком `CONTAINER_ID`.

Проверяем работу приложения:
```
docker exec -ti <CONTAINER_ID> bash
root@ef6a519b7e9c:/var/www/html# curl 127.0.0.1
...
                <h1>Simple PHP App</h1>
                <h2>Congratulations!</h2>
...
```

Ура! Первая сборка с помощью dapp прошла успешно.

Остановите контейнер (он должен удалиться после остановки, т.к. при его запуске была указана директива `--rm`):
```
docker stop <CONTAINER_ID>
```

Выполните очистку сборочного кэша dapp - это удалит образы на основе которых был собран образ приложения:
```
dapp dimg stages flush local
```

## Зачем нужен dapp?

Простое приложение показало, что dappfile может использоваться как замена Dockerfile. Но в чём же плюсы, кроме синтаксиса, немного похожего на Vagrantfile? Внутри dapp есть механизмы, которые незаметны на простом приложении, но для активно разрабатываемого приложения dapp может ускорить сборку и уменьшить размер финальных образов.

Узнать подробнее про возможности dapp можно по ссылками слева, либо продолжить ознакомление со списка возможностей ниже:

### Patch вместо полного копирования

В отличие от ADD и COPY dapp переносит изменённые файлы в образ с помощью патчей, а не передачей всех файлов ([подробнее](git.html)).

### Сборка образов по стадиям

dapp структурирует сборку, разбивая её на несколько стадий. Такое разбиение позволило ввести зависимости между сборкой стадии и изменениями файлов в репозитории. Например, сборка ассетов на стадии setup будет производиться, если изменились файлы в src, но более ранняя стадия install, где устанавливаются зависимости, будет пересобрана только, если изменился файл package.json ([подробнее](stages.html)).

### Несколько образов в одном Dappfile

Dapp умеет собирать сразу несколько образов по разному описанию в Dappfile. Подробнее в главе [сборка нескольких образов](multiple_images_for_build.html).

### Артефакты

Для уменьшения размера финального образа часто применяется способ использовать скачивание+удаление. Например так:

```
RUN “download-source && cmd && cmd2 && remove-source”
```

Dapp вводит в сборку понятие артефакта, такие вещи, как компиляция ассетов с помощью внешних инструментов, можно выполнить в другом контейнере и импортировать в финальный образ только нужные файлы ([подробнее](directives_artifact.html)).
