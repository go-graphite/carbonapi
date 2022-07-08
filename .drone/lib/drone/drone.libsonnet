local images = import 'images.libsonnet';
{
  step(name,
       commands,
       image=images._images.go,
       settings={},
       environment={},
       entrypoint=null,
       depends_on=[],
       dir=null):: {
    name: name,
    entrypoint: entrypoint,
    commands: if (dir == null || dir == '') then commands else ['cd %s' % dir] + commands,
    image: image,
    settings: settings,
    environment: environment,
    depends_on: depends_on,
  },

  withStep(step):: {
    steps+: [step],
  },

  withSteps(steps):: {
    steps+: steps,
  },


  withInlineStep(name,
                 commands,
                 image=images._images.go,
                 settings={},
                 environment={},
                 entrypoint=null,
                 depends_on=[],
                 dir=null)::
    $.withStep($.step(name, commands, image, settings, environment, entrypoint, depends_on, dir)),

  pipeline(name, steps=[], depends_on=null):: {
    kind: 'pipeline',
    type: 'docker',
    name: name,
    steps: steps,
    depends_on: depends_on,
  },
}