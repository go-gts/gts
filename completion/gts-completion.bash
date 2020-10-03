_gts_annotate()
{
    opts="-h --help --version -F --format -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_clear()
{
    opts="-h --help --version -F --format -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_complement()
{
    opts="-h --help --version -F --format -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_delete()
{
    opts="-h --help --version -F --format -e --erase -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_extract()
{
    opts="-h --help --version -F --format -m --range -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_insert()
{
    opts="-h --help --version -F --format -e --embed -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_length()
{
    opts="-h --help --version -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_query()
{
    opts="-h --help --version -d --delimiter --empty --no-header --no-key -n --name --no-location -o --output --source -t --separator"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_reverse()
{
    opts="-h --help --version -F --format -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_rotate()
{
    opts="-h --help --version -F --format -o --output -v --backward"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_search()
{
    opts="-h --help --version -F --format -e --exact -k --key --no-complement -o --output -q --qualifier"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_select()
{
    opts="-h --help --version -F --format -o --output -v --invert-match"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts_summary()
{
    opts="-h --help --version -F --no-feature -Q --no-qualifier -o --output"
    local cur="${COMP_WORDS[$COMP_CWORD]}"
    case "$cur" in
        -*)
            COMPREPLY=()
            while IFS='' read -r line
            do
                COMPREPLY+=("$line")
            done < <(compgen -W "$opts" -- "$cur")
            ;;
        *)
            COMPREPLY=()
            while IFS='' read -r line
            do 
                COMPREPLY+=("$line")
            done < <(compgen -f -- "$cur")
            ;;
    esac
}

_gts()
{
    cmds="-h --help --version annotate clear complement delete extract insert length query reverse rotate search select summary"
    local i=0 cmd

    while [[ "$i" -lt "$COMP_CWORD" ]]
    do
        local s="${COMP_WORDS[$i]}"
        case "$s" in
            gts)
                (( i++ ))
                break
                ;;
        esac
        (( i++ ))
    done

    while [[ "$i" -lt "$COMP_CWORD" ]]
    do
        local s="${COMP_WORDS[$i]}"
        case "$s" in
            -*) ;;
            *)
                cmd="$s"
                break
                ;;
        esac
        (( i++ ))
    done

    if [[ "$i" -eq "$COMP_CWORD" ]]
    then
        local cur="${COMP_WORDS[$COMP_CWORD]}"
        COMPREPLY=()
        while IFS='' read -r line
        do
            COMPREPLY+=("$line")
        done < <(compgen -W "$cmds" -- "$cur")
        return
    fi

    case "$cmd" in
        annotate)   _gts_annotate ;;
        clear)      _gts_clear ;;
        complement) _gts_complement ;;
        delete)     _gts_delete ;;
        extract)    _gts_extract ;;
        insert)     _gts_insert ;;
        length)     _gts_length ;;
        query)      _gts_query ;;
        reverse)    _gts_reverse ;;
        rotate)     _gts_rotate ;;
        search)     _gts_search ;;
        select)     _gts_select ;;
        summary)    _gts_summary ;;
        *) ;;
    esac
}

complete -F _gts gts