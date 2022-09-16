{
  secret(name, vault_path, key):: {
    kind: 'secret',
    name: name,
    get: {
      path: vault_path,
      name: key,
    },
  },
}