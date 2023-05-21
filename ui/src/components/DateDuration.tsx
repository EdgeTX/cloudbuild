import { Typography } from "antd";
const { Text } = Typography;

import dayjs from "dayjs";
import duration from "dayjs/plugin/duration";
import relativeTime from "dayjs/plugin/relativeTime";
dayjs.extend(duration);
dayjs.extend(relativeTime);

interface Props {
  lhsDate: string;
  rhsDate: string;
}

function DateDuration({ lhsDate, rhsDate }: Props) {
  const lhs = dayjs(lhsDate);
  const rhs = dayjs(rhsDate);

  const duration = dayjs.duration(rhs.diff(lhs));
  if (duration.asMilliseconds() <= 0) return null;

  let formatted = `${duration.seconds()} sec`;
  if (duration.minutes() > 0) 
    formatted = `${duration.minutes()} min ${formatted}`

  return <Text>{formatted}</Text>;
}

export default DateDuration;
