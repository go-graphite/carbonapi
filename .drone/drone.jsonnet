local drone = import 'lib/drone/drone.libsonnet';
local images = import 'lib/drone/images.libsonnet';
local triggers = import 'lib/drone/triggers.libsonnet';
local vault = import 'lib/vault/vault.libsonnet';

local pipeline = drone.pipeline;
local step = drone.step;
local withInlineStep = drone.withInlineStep;
local withStep = drone.withStep;
local withSteps = drone.withSteps;

local imagePullSecrets = { image_pull_secrets: ['dockerconfigjson'] };

local commentCoverageLintReport = {
  step: step('coverage + lint', $.commands, image=$.image, environment=$.environment),
  commands: [
    // Build drone utilities.
    'scripts/build-drone-utilities.sh',
    // Generate the raw coverage report.
    'go test -coverprofile=coverage.out ./...',
    // Process the raw coverage report.
    '.drone/coverage > coverage_report.out',
    // Generate the lint report.
    'scripts/generate-lint-report.sh',
    // Combine the reports.
    'cat coverage_report.out > report.out',
    'echo "" >> report.out',
    'cat lint.out >> report.out',
    // Submit the comment to GitHub.
    '.drone/ghcomment -id "Go coverage report:" -bodyfile report.out',
  ],
  environment: {
    GRAFANABOT_PAT: { from_secret: 'gh_token' },
  },
  image: images._images.goLint,
};

local buildAndPushImages = {
  // step builds the pipeline step to build and push a docker image
  step(app): step(
    '%s: build and push' % app,
    [],
    image=buildAndPushImages.pluginName,
    settings=buildAndPushImages.settings(app),
  ),

  pluginName: 'plugins/gcr',

  // settings generates the CI Pipeline step settings
  settings(app): {
    repo: $._repo(app),
    registry: $._registry,
    dockerfile: './Dockerfile',
    json_key: { from_secret: 'gcr_admin' },
    mirror: 'https://mirror.gcr.io',
    build_args: ['cmd=' + app],
  },

  // image generates the image for the given app
  image(app): $._registry + '/' + $._repo(app),

  _repo(app):: 'kubernetes-dev/' + app,
  _registry:: 'us.gcr.io',
};

local runTests = {
  step: step('run tests', $.commands, image=$.image),
  commands: [
    'make test'
  ],
  image: images._images.testRunner,
  settings: {

  }
};

[
  pipeline('test')
  + withStep(runTests.step)
  + imagePullSecrets
  + triggers.pr
  + triggers.main,

  pipeline('coverageLintReport')
  + withStep(commentCoverageLintReport.step)
  + triggers.pr,
]
+ [
  vault.secret('dockerconfigjson', 'secret/data/common/gcr', '.dockerconfigjson'),
  vault.secret('gh_token', 'infra/data/ci/github/grafanabot', 'pat'),
  vault.secret('gcr_admin', 'infra/data/ci/gcr-admin', 'service-account'),
  vault.secret('argo_token', 'infra/data/ci/argo-workflows/trigger-service-account', 'token'),
]