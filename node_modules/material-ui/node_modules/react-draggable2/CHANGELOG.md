# Changelog

### 0.1.0 (Jul 25, 2014)

- Initial release

### 0.1.1 (Jul 26, 2014)

- Fixing dragging not stopping on mouseup in some cases

### 0.2.0 (Sep 10, 2014)

- Adding support for snapping to a grid
- Adding support for specifying start position
- Ensure event handlers are destroyed on unmount
- Adding browserify support
- Adding bower support

### 0.2.1 (Sep 10, 2014)

- Exporting as ReactDraggable

### 0.3.0 (Oct 21, 2014)

- Adding support for touch devices

### 0.4.1 (Dec 7, 2014)

- Adding support for bounding movement

### 0.4.2 (Dec 7, 2014)

- Prevent errors when accessing browser globals

### 0.4.3 (Mar 2, 2015)

- Update dependencides
- Fix an issue where browser may be detected as touch-enabled but touch event isn't thrown. @STRML

### 0.5.0 (Mar 5, 2015)

- Remove dependency on Reactify for Browserify users
  [#2](https://github.com/mikepb/react-draggable/issues/2)
- Use Webpack directly for minification
- Source map files now have the `.js.map` file extention

### 0.5.1 (Mar 5, 2015)

- Remove `peerDependencies` from `package.json` #4
