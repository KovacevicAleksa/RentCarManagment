import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import AdminApp from "./components/AdminApp";
import { resolveAppMode } from "./lib/appMode";
import "./index.css";

const mode = resolveAppMode(import.meta.env.VITE_APP_MODE);
const Root = mode === "admin" ? AdminApp : App;

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>,
);
