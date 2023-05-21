import { useEffect, useState } from "react";

const QUERY = "(prefers-color-scheme: dark)";

function useColorScheme() {
  const getCurrentTheme = () => window.matchMedia(QUERY).matches;
  const [isDarkTheme, setIsDarkTheme] = useState(getCurrentTheme());

  useEffect(() => {
    const handleChange = () => {
      setIsDarkTheme(getCurrentTheme());
    };
    const matchMedia = window.matchMedia(QUERY);
    matchMedia.addEventListener("change", handleChange);
    return () => {
      matchMedia.removeEventListener("change", handleChange);
    };
  }, []);

  return isDarkTheme;
}

export { useColorScheme };
