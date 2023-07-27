# Entry point of the dcm command
dcm() {
  local OS=$(uname -s)
  local ARCH=$(uname -m)
  local BIN=$DCM_DIR/bin


  if [[ "$OS" == "Darwin" ]] && [[ "$ARCH" == "arm64" ]]; then
    BIN=$BIN/dcm-darwin-arm64
  elif [[ "$OS" == "Darwin" ]] && [[ "$ARCH" == "x86_64" ]]; then
    BIN=$BIN/dcm-darwin-amd64
  elif [[ "$OS" == "Darwin" ]] && [[ "$ARCH" == "arm64" ]]; then
    BIN=$BIN/dcm-darwin-amd64
  elif [[ "$OS" == "Linux" ]] && [[ "$ARCH" == "x86_64" ]]; then
    BIN=$BIN/dcm-linux-amd64
  elif [[ "$OS" == "FreeBSD" ]] && [[ "$ARCH" == "x86_64" ]]; then
    BIN=$BIN/dcm-freebsd-amd64
  elif [[ "$OS" == "CYGWIN_NT-6.1" ]] && [[ "$ARCH" == "x86_64" ]]; then
    BIN=$BIN/dcm-windows-amd64.exe
  else
    >&2 echo "Sorry, your OS ($OS) and Arch ($ARCH) is not currently supported by DCM." && \
        echo "Please submit your issue at https://github.com/beanworks/dcm/issues"
    return 1
  fi

  case "$1" in
    "goto" | "gt" | "cd" )
      cd $($BIN dir ${@:2})
      ;;
    "unload" | "ul" )
      unset -f dcm > /dev/null 2>&1
      unset DCM_DIR DCM_PROJECT > /dev/null 2>&1
      ;;
    * )
      $BIN ${@}
      ;;
  esac
}

# DCM command autocomplete handler
_dcm_complete() {
  local cur=${COMP_WORDS[COMP_CWORD]}
  local use=""

  case $COMP_CWORD in
    1)
      use="help setup run build shell purge branch goto update unload"
      ;;
    2)
      local prev_word=${COMP_WORDS[1]}
      case $prev_word in
        run|r)
          use="execute init build start stop restart up"
          ;;
        purge|rm)
          use="images containers all"
          ;;
        shell|sh|branch|br|goto|gt|cd|update|u)
          use=`dcm list`
          ;;
      esac
      ;;
  esac

  COMPREPLY=( $( compgen -W "$use" -- $cur ) )
}

complete -o default -F _dcm_complete dcm
