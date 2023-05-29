import Layout from "@comps/Layout";
import { Decorator } from "@storybook/react";

const withPage: Decorator = (Story) => (
  <Layout>
    <Story />
  </Layout>
);

export { withPage };
