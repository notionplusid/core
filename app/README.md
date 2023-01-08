# Plus ID

The Plus ID component is designed to be ran on GCP's App Engine. If the default configurations are used, the resource usage would remain within the GCP's free tier, thus your monthly expense should remain ~$0.00.

## Deployment to private repo.
### 0. Apply `tf` configuration
Follow the [Terraform configuration](../tf/README.md) instruction if not done before.

### 1. Initialise the configuration
Use the configuration template by
```bash
cp app.yaml.example app.yaml
```

### 2. Change the configuration
Change the value of `env_variables.GCLOUD_PROJECT_ID` to your GCP project ID.

> The `us-east2` region is used to be as close as possible to the Notion's region. This way, the delay in calling the Notion API should be minimal.

### 3. Deploy to App Engine
Simply run
```bash
gcloud app deploy --version v1
```

And follow the instructions.

### 4. Set up your Notion database
Follow up the [instructions](https://notionplusid.app/welcome) on the website