dcm() {
  DCM_BIN=$DCM_DIR/bin/dcm
  [ "$1" == "" ] && echo "123"
  case "$1" in
    "goto" | "gt" | "cd" )
      cd $($DCM_BIN dir ${@:2})
      ;;
    * )
      $DCM_BIN ${@}
      ;;
  esac
}
