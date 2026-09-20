import { app as n, BrowserWindow as s } from "electron";
import { fileURLToPath as a } from "node:url";
import o from "node:path";
const t = o.dirname(a(import.meta.url));
process.env.APP_ROOT = o.join(t, "..");
const i = process.env.VITE_DEV_SERVER_URL, m = o.join(process.env.APP_ROOT, "dist-electron"), r = o.join(process.env.APP_ROOT, "dist");
process.env.VITE_PUBLIC = i ? o.join(process.env.APP_ROOT, "public") : r;
let e;
function l() {
  e = new s({
    width: 1020,
    height: 720,
    icon: o.join(process.env.VITE_PUBLIC, "skydock-logo-resized.png"),
    maximizable: !1,
    resizable: !1,
    webPreferences: {
      preload: o.join(t, "preload.mjs")
    }
  }), e.webContents.on("did-finish-load", () => {
    e == null || e.webContents.send("main-process-message", (/* @__PURE__ */ new Date()).toLocaleString());
  }), i ? e.loadURL(i) : e.loadFile(o.join(r, "index.html"));
}
n.on("window-all-closed", () => {
  process.platform !== "darwin" && (n.quit(), e = null);
});
n.on("activate", () => {
  s.getAllWindows().length === 0 && l();
});
n.whenReady().then(l);
export {
  m as MAIN_DIST,
  r as RENDERER_DIST,
  i as VITE_DEV_SERVER_URL
};
