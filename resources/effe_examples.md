# Effe examples

## First example

```golang
func (c *CreateUserService) Do(ctx context.Context, user User) error {
    err := c.validUser(ctx, user)
    if err != nil {
        return errors.Wrap(err, "something goes wrong: validUser")
    }

    savedUser, err := c.saveUser(ctx, user)
    if err != nil {
        return errors.Wrap(err, "something goes wrong: saveUser")
    }

    err = c.notifyUser(ctx, user)
    if err != nil {
        return errors.Wrap(err, "something goes wrong: notifyUser")
    }
    return nil
}
```

## Second example

```golang
func (c *CreateUserService) Do(ctx context.Context, user User) (err error) {
    defer func() {
        anaErr := c.sendAnalyticsEvent(ctx, user)
        if anaErr != nil && err != nil {
            err = errors.Wrap(err, anaErr.Error())
        }
        //...
        //...
    }()
    err = c.validUser(ctx, user)
    if err != nil {
        return errors.Wrap(err, "something goes wrong: validUser")
    }

    savedUser, err := c.saveUser(ctx, user)
    if err != nil {
        return errors.Wrap(err, "something goes wrong: saveUser")
    }

    err = c.notifyUser(ctx, user)
    if err != nil {
        return errors.Wrap(err, "something goes wrong: notifyUser")
    }
    return nil
}
```

## Third example

```golang
func (c *CreateUserService) Do(ctx context.Context, user User) error {
    err := func(ctx context.Context, user User) {
        err = c.validUser(ctx, user)
        if err != nil {
            return errors.Wrap(err, "something goes wrong: validUser")
        }

        savedUser, err := c.saveUser(ctx, user)
        if err != nil {
            return errors.Wrap(err, "something goes wrong: saveUser")
        }

        err = c.notifyUser(ctx, user)
        if err != nil {
            return errors.Wrap(err, "something goes wrong: notifyUser")
        }
        return nil
    }(ctx, user)
    anaErr := c.sendAnalyticsEvent(ctx, user)
    if anaErr != nil && err != nil {
        err = errors.Wrap(err, anaErr.Error())
    }
    //...
    //...
    return nil
}
```
