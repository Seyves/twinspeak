// Allow standard CSS imports
declare module '*.css'

// If you are using CSS Modules
declare module '*.module.css' {
    const classes: { [key: string]: string }
    export default classes
}
