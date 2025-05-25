# Common CRUD PoC Explanation

As background, I would recommend reading [Getting started with generics] if you are not already familiar with generics. The next section will explain generics, although focusing more on how this PoC uses them.

[Getting started with generics]: https://go.dev/doc/tutorial/generics

## Generics in this PoC

### What are Generics?

Generics introduce the concept of type parameters, which are similar to function parameters but rather than accept values of a given type, they accept a type itself. This allows the creation of structs, interfaces, and functions that may be used with many different types, but behave as though they are written for a single type.

The following example should illustrate the basics of calling generic functions with type parameters.

```go
// This function is equivalent to the builtin [new]. The `any` after the type parameter `T` is a constraint.
func CreateNew[T any]() *T {
    return new(T)
}

func CreateNewIntWithoutGenerics() *int {
    return new(int)
}

func CreateNewFloat64Withoutgenerics() *float64 {
    return new(float64)
}

func main() {
    // adding string inside the brackets sets the type parameter T to string
    // this would be equivalent to a function with signature func CreateNew() *string
    var myStringPointer *string = CreateNew[string]()

    // equivalent to
    //  var myIntPointer *int = CreateNew[int]()
    var myIntPointer *int = CreateNewIntWithoutGenerics()

    // equivalent to
    //  var myFloat64Pointer *float64 = CreateNew[float64]()
    var myFloat64Pointer *float64 = CreateNewFloat64Withoutgenerics()
}
```

### Generic Constraints

The above example used `any` as a constraint, which allows the caller to provide any type as a type parameter. Constraints are just interfaces, however. `any`, for example, is a builtin type alias for `interface{}`. Furthermore, these interfaces can themselves be generic.

Constrains are different from just providing an interface, however. In the example, although `T` has constraint `any`, the actual type that the function returns is not `*any`, but the pointer to the type that implements `any`.

Interface syntax has also been extended to support specifying the underlying type that implements an interface:

```go
// The tilde means that any type derived from an int fulfills [Integer]. If the tilde were not added, only `int` would
// implement this interface.
type Integer interface {
    ~int
}

// Because of the tilde, the [MyInt] type implements Integer even though it is a distinct type from `int`.
type MyInt int

// This is a generic interface that takes a type parameter `T` that can be any type which implements [Integer]. It is
// not implemented by any [Integer] but rather any pointer to a type that implements [Integer]. For example, MyInt
// does not implement IntegerPointer but `*MyInt` implements `IntegerPointer[MyInt]`.
type IntegerPointer[T Integer] interface {
    *T
}

// Here we use define a generic struct that has two different type parameters, one of which depends on the other. We can
// use it to store a pointer to MyInt, in which case the type would be `IntegerPointerContainer[MyInt, *MyInt]`. This
// combines the fact that MyInt implements Integer and can be used as T and *MyInt implements IntegerPointer[MyInt] and
// can be used as ST (star T, hence ST).
type IntegerPointerContainer[T Integer, ST IntegerPointer[T]] struct {
    storedInteger ST
}
```

The last type, `IntegerPointerContainter` is particularly useful for understanding this PoC as the pattern is used multiple times. It allows for taking a type parameter while adding a constraint on its pointer.

### objectPointer

When creating a generic version of `Get`, it is necessary to create a pointer to the Kubernetes resource struct. If we wish to implement the function for a `corev1.Namespace`, for example, we would have to write the function like this:

```go
// ...inside the Get function...
var namespace *corev1.Namespace = new(corev1.Namespace)
err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKeyFromObject(builder.Definition), namespace)
```

A generic version would instead use some generic type instead of `corev1.Namespace`:

```go
// ...inside the Get function...
var object *T = new(T)
err := builder.apiClient.Get(context.TODO(), runtimeclient.ObjectKeyFromObject(builder.Definition), object)
```

This of course requires knowing the type parameter `T`, presumably a struct. Knowing just the pointer type `*T` would not let us write the function, since `new(*T)` would be type `**T` and dereferencing gives `*T` which is also `nil`.

What is not immediately obvious here is that the generic version works only if `object` implements `runtimeclient.Object`. So now we need to ensure that we have a type parameter `T` and its pointer, `*T`, implements `runtimeclient.Object`.

This is why this PoC defines the `objectPointer` interface to use as a constraint for the builder and its methods.

```go
type objectPointer[T any] interface {
    *T
    runtimeclient.Object
}

// Now we can define the Get function signature with the required constraints
func Get[T any, ST objectPointer[T]](builder Builder[T, ST]) *T
```

## The Builder Interface

The `Builder` interface is intentionally placed under `pkg/internal` since it is not designed for use outside of this library. It is useful for creating common functions, but does not have much value for consumers of the library where the individual structs that implement `Builder` are much more useful.

Similarly, the functions in `Builder` are only there as utilities for creating common functions. As a result of Go's visibility rules, the methods will still appear to users of `eco-goinfra`, but there is no reason to use them.

Knowing the reason for the `objectPointer`, there is not all that much special about `Builder` other than the use of `ST` in the returns of its methods. Although they are meant to be implemented as returning `*T`, `ST` must be explicitly used so that Go can infer the return types implement `runtimeclient.Object`.

### CRUD methods

The usual CRUD methods on builders (`Create`, `Get`, `Exists`, `Update`, and `Delete`) are implemented in `common.go` as functions that take a `Builder`. Since methods cannot be defined with interfaces as receivers, these must be functions instead. Additionally, the type constraints must be repeated on each function.

The CRUD functions are implemented similarly to how they would be on an existing builder. However, they must use exclusively the functions in the `Builder` interface and other common functions. For example, `builder.validate()` becomes `Validate(builder)`.

One small difference from usual is in `Get`, where contrary to the example from above, we must use `var object ST = new(T)`. This is again a limitation of Go where `object := new(T)` would be type `*T` which is not inferred to implement `runtimeclient.Object`.

Logging is the biggest difference and something that this PoC does not address very well. Since we do not know if a resource is namespaced inside the function, we have to check if the definition includes a namespace and adjust logging based on that. The name of the resource must also be replaced with a value based on the GVK's kind.

### Validate

Something that Brandon figured out with his PoC is that comparing to `nil` is not always sufficient for ensuring a value is not nil. When working with pointers directly, this will not come up, but since we now have to deal with the builder being an interface as well as the client being an interface, it must be addressed.

This section can mostly be skipped if you just follow the rule that for interfaces, you need to write `iface == nil || reflect.ValueOf(iface).IsNil()`.

In Go, interface values are not the same as the type that implements an interface. If you are more familiar with Rust, interface values align with trait objects and can be thought of as `Box<dyn Trait>`. Sticking to Go, an interface value is more akin to a struct like this:

```go
// Note that this is not meant to be an accurate description of Go internals, just useful for thinking about them
type iface struct {
    value unsafe.Pointer // or void* in C
    typ reflect.Type
}
```

So when you assign a concrete type to an interface value like the following, Go implicitly does the conversion.

```go
var runtimeObject runtime.Object = &corev1.Namespace{}
// implicitly Go does something like this:
var runtimeObject *iface = &iface{value: &corev1.Namespace, typ: corev1.Namespace}
```

This implicit conversion is why the following does not work:

```go
var runtimeObjects []runtime.Object = []*corev1.Namespace{new(corev1.Namespace)}
```

Converting a single concrete type into its interface is something Go does implicitly, but converting a slice fo concrete types into a slice of interfaces is different. It would require allocating a whole new slice then converting each value into an interface value and adding it to the new slice.

This difference between the underlying value and the implicit interface value is why we must do two different checks for nil. Comparing against nil directly only ensures that the interface value itself is not nil. Using reflect is what checks whether the concrete value (the `value` field in the examples) is not nil.

You should avoid creating situations where the interface is not nil but the concrete value is, although it can be easy to do unintentionally. `Validate(nil)` will create a nil interface value that can be compared directly. The following requires reflect to properly validate:

```go
var namespaceBuilder *namespace.Builder // nil by default
Validate(namespaceBuilder) // implicit conversion of *namespace.Builder to Builder[corev1.Namespace, *corev1.Namespace]
```

### Implementing Builder

My original idea was that we can wrap these common functions in the actual builders. [clusterinstance.go](../../siteconfig/clusterinstance.go) in this PoC has been updated to work like this. However, this requires a lot of boilerplate and to Trey's point, it would cause issues for our unit test coverage numbers.

The flexibility this style of using `Builder` grants us can be useful for making generic versions of other methods, or for methods where the existing builders do not all use the same signature (`Update` may or may not take a parameter, for example).

## The EmbeddableBuilder Struct

Taking inspiration from Trey's PoC, the EmbeddableBuilder allows us to implement `Builder` and the CRUD methods on a single struct that gets embedded inside existing builders. Just by defining the `NewBuilder`, `Pull`, and `GetKind` functions, we can create a new builder that has all of the CRUD methods in 20 lines or so. [placementrule.go](../../ocm/placementrule.go) provides an example in this PoC, although not with all of the CRUD methods.

`EmbeddableBuilder` itself looks very similar to existing builders, except that we have the same generic parameters as `Builder` and use `ST` for the definition and object. The `apiClient` is also constrained to `runtimeclient.Client`, although this is similar to most builders.

Implementing `Builder` is mostly a matter of providing getters and setters for each field. `GetKind` is the exception, which returns a zero value for reasons discussed in the next section.

Once `Builder` is implemented, CRUD methods are just a thin wrapper on the common functions:

```go
func (builder *EmbeddableBuilder[T, ST]) Get() (*T, error) {
    return Get(builder)
}
```

### GetKind

The `Builder` interface requires that implementors have a `GetKind` method, with the additional stipulation that it should function as intended even on the zero value of a builder. However, the zero value of `EmbeddableBuilder` does not have any way to get the kind.

`metav1.TypeMeta` provides a `GetObjectKind` method, but for a zero value of a resource, this will also return a zero value. The way this works for controller-runtime is through schemes, where the processes of registering a new type creates a mapping between its Go type and its GVK. The zero value of `EmbeddableBuilder` does not have the scheme though.

The solution (unless someone can think of a better one) is to return the zero value for `EmbeddableBuilder` itself then require individual builders override this embedded method with their own that returns a constant GVK. Since we really care about builders which embed `EmbeddableBuilder` implementing `Builder`, less so `EmbeddableBuilder` itself, this system of overriding works transparently.

One may still access the embedded version using `builder.EmbeddableBuilder.GetKind`, but there is no reason to do so.

Something that this PoC does not address but must be worked out a bit more is that `GetKind` will have the zero value inside of implementations of CRUD methods for `EmbeddableBuilder`. I have some ideas but have not decided the best way to address this.

### NewBuilder and Pull Functions

For the purposes of using `EmbeddableBuilder`, the NewBuilder and Pull functions operate the same way.

In the PoC, each starts with a check that `apiClient` is nil before passing `apiClient.Client` to the common version of the function, although this is likely unnecessary since `clients.Settings` embeds `runtimeclient.Client`.

These each call common functions with the same signature, although with double the type parameters. This is since we must create a new pointer to the builder struct and run into a similar scenario that requires `objectPointer`. It is a little different since the pointer must implement a generic interface but the purpose is the same.

The ordering of the generic parameters allows us to only specify the non-pointer types for the Kubernetes resource and its builder since the pointers can be inferred. This inference only happens on function calls, not with type instantiations, which is why all parameters must appear in those cases.

The last unique thing about these functions is that they take scheme attachers as parameters. These must be provided explicitly, but since it is the same for a given builder, users of `eco-goinfra` will not have to worry about it.

#### Non-trivial NewBuilder Functions

PlacementRule was chosen for this PoC since it does not have any other parameters in the NewBuilder function. In cases where there are additional parameters, they must be validated after calling the common function. Otherwise, not much will change.

```go
// ...inside NewBuilder...
builder := common.NewNamespacedBuilder[placementrulev1.PlacementRule, PlacementRuleBuilder](
    apiClient.Client, placementrulev1.AddToScheme, name, nsname)

if builder == nil {
    return nil
}

if builder.GetErrorMessage() == "" {
    return builder
}

if myOtherParameter == "" {
    builder.SetErrorMessage("myOtherParameter cannot be empty")

    return builder
}

return builder
```

I will need to work on making this a bit easier to use, but the general idea should remain the same.

### Extending with EmbeddableBuilder

Adding extra methods to builders that embed `EmbeddableBuilder` works almost identically as without it. However, the error message and apiClient are still private so the corresponding getters and setters must be used instead.

If we have getters and setters anyway, there's not that much difference to just making the fields public, but hopefully their use is discouraged by making them method calls.

We may also find that there are other methods beyond CRUD ones that could be made generic for more builders. It would be nice to keep the `EmbeddableBuilder` constrained to CRUD methods, but new methods can still be added to builders that wrap common functions. `DeleteAndWait` could be implemented this way, for example.

Even more methods could be made generic by taking an accessor function as a parameter, similar to how NewBuilder relies on being provided a scheme attacher function.

## Appendix: Assertions

In cases where we want to ensure that a struct (or pointer to struct) implements an interface, we can use a variable declaration as a compile time assertion. The ClusterInstance in this PoC is one example:

```go
var _ common.Builder[
    siteconfigv1alpha1.ClusterInstance, *siteconfigv1alpha1.ClusterInstance] = (*CIBuilder)(nil)
```

All this does is ensure we can assign a `*CIBuilder` to its corresponding `Builder` interface. It has no meaning at runtime and I would suspect it gets optimized out by the compiler, although I have done nothing to confirm this.

I got carried away with EmbeddableBuilder to create a similar assertion for a generic struct. It would probably be better and just as useful to stick to using something like `corev1.Namespace` as a dummy type to ensure the assertion is satisfied.
