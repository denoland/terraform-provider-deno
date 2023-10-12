const keys = [
  "TF_ACC",
  "DENO_DEPLOY_ORGANIZATION_ID",
  "DEPLOY_API_HOST",
  "DENO_DEPLOY_TOKEN",
];

for (const key of keys) {
  console.log(`${key}: ${Deno.env.get(key)}`);
}
