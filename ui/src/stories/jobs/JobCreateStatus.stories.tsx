import JobCreateStatus from "@comps/jobs/JobCreateStatus";
import type { Meta, StoryObj } from "@storybook/react";

const meta: Meta = {
  title: "jobs/JobCreateStatus",
  component: JobCreateStatus,
};

export default meta;
type Story = StoryObj;

export const Success: Story = {
  args: {
    jobStatus: {
      job: {
        id: "7ea44f17-66ce-499c-b72b-0f320231eeff",
        status: "WAITING_FOR_BUILD",
        release: "nightly",
        target: "xlite",
      },
      params: {
        release: "nightly",
        target: "xlite",
        flags: [],
      },
    },
  },
};

export const AlreadyBuilt: Story = {
  args: {
    jobStatus: {
      job: {
        id: "bfd112ce-02a6-41fe-8414-41e705c79aa8",
        status: "BUILD_SUCCESS",
        release: "nightly",
        target: "boxer",
      },
      params: {
        release: "nightly",
        target: "boxer",
        flags: [],
      },
    },
  },
};

export const Error: Story = {
  args: {
    jobStatus: { "error": "release not supported" },
  },
};
