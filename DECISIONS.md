# Decisions

## Contract-check baseline (`contract-baseline.json`)

The SDK predates the drift gate and already lacks parts of the public spec
(models and endpoints it never implemented). The baseline lists those known
gaps so the gate fails only on new drift. It should shrink as coverage lands;
entries that stop being missing fail CI as stale.

## Hand-listed registry (`cmd/dumpsurface/registry.go`)

Go reflection needs compile-time types and cannot enumerate a package, so the
exported structs are listed by hand. `TestContractSurface_RegistryCoversPackage`
fails when a struct is added without being registered.
