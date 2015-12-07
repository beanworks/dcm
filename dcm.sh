dcm() {
  DCM_BIN=$DCM_DIR/bin/dcm
  case "$1" in
    "goto" | "gt" | "cd" )
      cd $($DCM_BIN dir ${@:2})
      ;;
    "unload" | "ul" )
      unset -f dcm > /dev/null 2>&1
      unset DCM_DIR DCM_PROJECT > /dev/null 2>&1
      ;;
    * )
      $DCM_BIN ${@}
      ;;
  esac
}
