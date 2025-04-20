import type { CodegenConfig } from '@graphql-codegen/cli';

const config: CodegenConfig = {
  overwrite: true,
  schema: "../graph/schema.graphqls",
  generates: {
    "lib/gql/": {
      preset: 'client',
      documents: '*.graphql',
      config: {
        documentMode: 'string'
      }
    }
  }
};

export default config;