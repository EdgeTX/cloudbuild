import type { Meta, StoryObj } from "@storybook/react";
import JobCreation from "@pages/JobCreation";
import { withPage } from "../decorators/withPage";
import { withProvider } from "../decorators/withProvider";

const meta: Meta = {
  title: "pages/JobCreation",
  component: JobCreation,
  decorators: [withPage, withProvider],
};

export default meta;
type Story = StoryObj;

export const Primary: Story = {
  args: {},
};
