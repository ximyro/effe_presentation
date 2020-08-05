---
theme: base
paginate: true
marp: false
auto-scaling: true
---

<style>
@import url('https://fonts.googleapis.com/css?family=Proxima+Nova');
section {
    font-family: 'proxima-nova', sans-serif;
    background: rgb(249, 182, 27)
}

section h1{
    font-size: 72pt;
}

section h1, section h2{
    color: black;
}

section h2 {
  font-size: 22pt;
}

section.list ul li {
  margin-left: 70%;
}

section.half {
  background: linear-gradient(90deg, #f9b61b 50%, #FFF 50%);
}
</style>

# Business logic without pain (Go edition)

Buzlov Ilya | 2020.04.20

---

## Plan

<!-- _class: half list -->

* Foo
* Fo
* Fo

---

## GoLang норм

---

## Терминология

- Шаг - Действие совершаемое в процессе работы сценария
- Сценарий - Последовательность выполнения действий

---

## Очень очень очень много кода в Golang

---

## Как много

```go
 user, err := createUser()
 if err != nil {
   return errors.Wrap(err, "can't create user")
 }

```

- 2 шага + обработка ошибок ~ 8 строк кода
- 40 шагов + обработка ошибок ~ 160 строк кода

---
20 шагов + обработка ошибок ~ 300 строк кода

---

## Внятность

Сложно понять как работает весь сценарий целиком:

- Бизнес-логика разбита по пакетам
- Бизнес-логики много

<!-- Сложно делить бизнес-логику -->
<!--
Представим вы зашли в проект, и вам надо понять как работает какой-то там флоу?
Как написать бизнес-логику так, чтобы взглянув на какую-то часть кода вы могли понять хотя бы схематично как он работает?
Если она разбита по пакетам, вы идете прыгаете от пакета к пакету и в какой-то момент уже не понимаете где находитесь и где были. И как данные передаются между ними.
-->

---

## Расширяемость

Что сделать, чтобы добавить новый шаг в сценарий

- Написать сам шаг
- Найти место куда его вставить
- Добавить зависимости для конкретного шага в сервис объект
- Написать тест для шага
- Написать тест для флоу

---

## Все понятно

- Написать сам шаг
- Написать тест для шага

---

## Найти место куда его вставить

- Сложно
- Можно что-то забыть/не заметить
- Добавить зависимости для конкретного шага в сервис объект

---

## Написать тест для флоу

Проинициализовать флоу со всеми зависимостями, чтобы написать тест

- Громоздкие тесты
- Неудобно с этим работать

---

## Переиспользуемость

Представим что у нас есть флоу:

- buildUserFromRequest
- createUser
- sendNotification

---

и нам сказали добавить еще один flow

- buildUserFromRequest
- **findUserByID**
- **updateUser**
- sendNotification

---

## Какие тут есть проблемы?

Делать один сценарий или два?

---

## Если сделаем один

Плюсы

- Меньше кода
- Быстрее делать

Минусы

- Сложнее расширять
- Сложнее отлаживать
- Риск сломать сразу оба сценацрия

---

## Что если сделаем два

Плюсы

- Легче расширять
- Легче отлаживать
- Риск сломать что-то меньше

Минусы

- Больше кода, сложнее поддержка
- Больше времени на разработку

---

## Усложним

- buildUserFromRequest
- validateUser
- createUser
- createCustomer
- sendNotification

---

- buildUserFromRequest
- validateUser
- findUserByID
- updateUser
- updateCustomer
- sendNotification

---

## Много ручной работы

- Написать последовательность вызовов
- Для каждого шага надо передать нужные ему зависимости
- Добавить зависимость для каждого шага в сервис объект
- Думать над враппингом ошибок

---

## Effe

- Provide visibility and traceability into these process flows
- Errors are wrapping automatically
- Dependencies build for flow automatically
- Easy flow debugging
- Easy flow extending
- Split dependencies for steps in flow: the step has only dependencies that it needs
- Allow greater reuse of existing functions
- Easy flow testing: small interface/easier to understanding what happening

---

```go
func buildUser() func(UserAttributes) User {
    return func(uAttrs UserAttributes) User {
        return User{
            Email:    uAttrs.Email,
            Password: uAttrs.Password,
        }
    }
}

func createUser(userRepo UserRepository) func(context.Context, User) error {
    return func(ctx context.Context, user User) error {
        return userRepo.Create(ctx, user)
    }
}
```

---
## effe.go

```go
// +build effeinject

package actions

import (
    "github.com/gtforge/paymentsos_integration_service/pkg/effe"
)

func BuildCreateUserFlow(uAttrs UserAttributes) error {
    effe.BuildFlow(
        effe.Step(buildUser),
        effe.Step(createUser),
    )
    return nil
}
```

---
## effe_gen.go

```go
func BuildCreateUserFlow(service BuildCreateUserFlowService) BuildCreateUserFlowFunc {
    return func( ctx context.Context, uAttrs UserAttributes) error {
        UserVal := service.BuildUser(uAttrs)
        err := service.CreateUser(ctx, UserVal)
        if err != nil {
            return errors.Wrap(err, "failed createUser")
        }
    return nil
    }
}

```

---

## Сервис-объект

```go
type BuildCreateUserFlowService interface {
    BuildUser(uAttrs UserAttributes) User
    CreateUser(ctx context.Context, user User) error
}
```

---

## Тестирование

```go
serviceMock := mocks.NewMockBuildCallbackFlowService(ctrl)
callbackHandleFunc := BuildCallbackFlow(serviceMock)
serviceMock.EXPECT().BuildUser(uAttrs).Return(User{})
serviceMock.EXPECT().CreateUser(User).Return(nil)
err := callbackHandleFunc(ctx, zoozRequest)
assert.NoError(t, err)


```

---

## Имплементация

```go
type BuildCreateUserFlowImpl struct {
    buildUserFieldFunc  func(uAttrs UserAttributes) User
    createUserFieldFunc func(ctx context.Context, user User) error
}
```

---

## Инициализация зависимостей

```go
func NewBuildCreateUserFlowImpl(userRepo UserRepository) *BuildCreateUserFlowImpl {
    return &BuildCreateUserFlowImpl{buildUserFieldFunc: buildUser(), createUserFieldFunc: createUser(userRepo)}
}
```

---

## API

- Step
- Wrap
- Decision
- Failure

---

## Step

```go
 effe.BuildFlow(
    effe.Step(buildUser),
    effe.Step(createUser),
)
```

---

## Wrap

```go
func BuildCreateUserFlow(uAttrs UserAttributes) error {
    effe.BuildFlow(
        //...,
        effe.Wrap(effe.Before(beginTransaction), effe.Success(commitTransaction), effe.Failure(rollbackTransaction),
            effe.Step(createUser),
        )
        //...,
    )
    return nil
}
```

---

```golang
//...
err16 := func(ctx context.Context, user User) error {
  transaction err15 := service.BeginTransaction(ctx)
  if err15 != nil {
    err15 = service.rollbackTransaction(ctx, err15, transaction)
    return errors.Wrap(err15, "failed beginTransaction")
  }

  err14 := service.CreateUser(ctx, user)
  if err != nil {
    err14 = service.rollbackTransaction(ctx, err15, transaction)
    return errors.Wrap(err14, "failed createUser")
  }

  err13 := service.CommitTransaction(transaction)
  if err != nil {
    err13 = service.rollbackTransaction(ctx, err13, transaction)
    return  errors.Wrap(err13, "failed commitTransaction")
  }
  return nil
}
//...
```

---

## Decision

```go
func BuildCreateUserFlow(uAttrs UserAttributes) error {
    effe.BuildFlow(
        //....,
            effe.Decision(new(entities.LockCreated),
                effe.Case(false, effe.Step(stop())),
                effe.Case(true,
                    effe.Step(createUser),
                    effe.Step(createCustomer),
                ),
        //....,
    )
    return nil
)
```

---

```go
//...
err15 = func(lockCreatedVal entities.LockCreated, user User) error {
   switch lockCreatedVal {
      case true:
        err := func(ctx context.Context) {
          err = createUser(user)
          if err != nil {
            return errors.Wrap("foo")
          }
          err = createCustomer(user)
          if err != nil {
            return errors.Wrap("foo")
          }
        }()
      case false:
        service.Stop()
        return nil
      default:
        return errors.New("unsupported type lockCreatedVal")
   }
}
//...
```

---

## Failure

- BuildFlow
- Wrap
- Decision

---

```go
func BuildCreateUserFlow(uAttrs UserAttributes) error {
    effe.BuildFlow(
        effe.Step(buildUser),
        effe.Step(createUser),
        effe.Failure(handleErr),
    )
    return nil
)
```

---

```go
func BuildCreateUserFlow(service BuildCreateUserFlowService) BuildCreateUserFlowFunc {
    return func( ctx context.Context, uAttrs UserAttributes) error {
        UserVal := service.BuildUser(uAttrs)
        err := service.CreateUser(ctx, UserVal)
        if err != nil {
            err = handleErr(err)
            return errors.Wrap(err, "failed createUser")
        }
    return nil
    }
}
```

---

## Failure in Wrap

```go
func BuildCreateUserFlow(uAttrs UserAttributes) error {
    effe.BuildFlow(
        //...,
        effe.Wrap(effe.Before(beginTransaction), effe.Success(commitTransaction), effe.Failure(rollbackTransaction),
            effe.Step(createUser),
            effe.Step(createUserPermissions),
            effe.Step(createUserSettings),
        )
        //...,
    )
    return nil
}
```

---

```go
func BuildCallbackFlow(zoozReq entities.ZoozRequest) error {
  effe.BuildFlow(
    effe.Step(withWaitTime),
    effe.Wrap(effe.Before(createExecutionZoozRequestLock), effe.Success(unlock), effe.Failure(catchErrAndUnlock),
      effe.Decision(new(entities.LockCreated),
        effe.Case(false, effe.Step(stop)),
        effe.Case(true,
          effe.Step(findRequestByZoozRequest),
          effe.Step(checkRequestIsFound),
          effe.Decision(new(entities.RequestFound),
            effe.Case(false, effe.Step(stop)),
            effe.Case(true,
              effe.Wrap(effe.Before(createExecutionRequestLock), effe.Success(unlock), effe.Failure(catchErrAndUnlock),
                effe.Decision(new(entities.LockCreated),
                  effe.Case(false, effe.Step(stop)),
                  effe.Case(true,
                    effe.Step(calcCheckRequestStateCases),
                    effe.Decision(entities.CheckRequestStateCases{}.Case,
                      effe.Case("write", effe.Step(writeGPMResult)),
                      effe.Case("stop", effe.Step(stop)),
                      effe.Case("assign",
                        effe.Step(checkIsProviderError),
                        effe.Decision(new(entities.RetryableFailedZoozRequest),
                          effe.Case(true, effe.Step(removeTransactionID), effe.Step(createRequestJob)),
                          effe.Case(false,
                            effe.Step(assignRequestState),
                            effe.Decision(entities.Request{}.State,
                              effe.Case(entities.SuccessedState, effe.Step(writeGPMResult)),
                              effe.Case(entities.FailedState, effe.Step(writeGPMResult)),
                              effe.Case(entities.PendingState, effe.Step(stop)),
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    ),
    effe.Failure(failure),
  )
  return nil
}
```

---

## Кастомизация

```golang
package myeffe

import (
  "github.com/gtforge/paymentsos_integration_service/pkg/effe"
)

type conditionComponent struct {

}

func genConditionComponent(effeConditionCall *ast.CallExpr, f *effe.FlowGen) (effe.Component, error) {
  return &conditionComponent{}, nil
}


func main() {
  effe.Register("IfCond", genConditionComponent)
  effe.Run()
}

```

---

## Что дальше?

- Генерация диаграмм/документации
- Трейсинг сценария(input/output для каждого шага) + схема
- Работа над синтаксисом/генерацией
- Новые API: Rescue/Try, Retry
