import localFont from "next/font/local";


/**
 * The Inter font from the local filesystem.
 */
export const Inter = localFont({
    src: "../node_modules/inter-ui/variable/InterVariable.woff2",
    variable: "--sans",
});
