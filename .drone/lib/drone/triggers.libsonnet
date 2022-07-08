{
  main:: {
    trigger+: {
      branch+: ['main'],
      event+: {
        include+: ['push'],
      },
    },
  },
  pr:: {
    trigger+: {
      event+: {
        include+: ['pull_request'],
      },
    },
  },
  // excluding paths disables runs that contain changes to ONLY these files
  excludeModifiedPaths(paths):: {
    trigger+: {
      paths+: {
        exclude+: paths,
      },
    },
  },
}