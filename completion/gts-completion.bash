_gts_annotate()
{
    opts="-h --help --version -F --format --no-cache -o --output"
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

_gts_cache_list()
{
    opts="-h --help --version "
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

_gts_cache_path()
{
    opts="-h --help --version "
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

_gts_cache_purge()
{
    opts="-h --help --version "
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

_gts_cache()
{
    cmds="-h --help --version list path purge"
    local i=0 cmd

    while [[ "$i" -lt "$COMP_CWORD" ]]
    do
        local s="${COMP_WORDS[$i]}"
        case "$s" in
            cache)
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
        list)  _gts_cache_list ;;
        path)  _gts_cache_path ;;
        purge) _gts_cache_purge ;;
        *) ;;
    esac
}

_gts_clear()
{
    opts="-h --help --version -F --format --no-cache -o --output"
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
    opts="-h --help --version -F --format --no-cache -o --output"
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

_gts_define()
{
    opts="-h --help --version -F --format --no-cache -o --output -q --qualifier"
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
    opts="-h --help --version -e --erase -F --format --no-cache -o --output"
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
    opts="-h --help --version -F --format -m --range --no-cache -o --output"
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

_gts_infix()
{
    opts="-h --help --version -e --embed -F --format --no-cache -o --output"
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
    opts="-h --help --version -e --embed -F --format --no-cache -o --output"
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

_gts_join()
{
    opts="-h --help --version -c --circular -F --format --no-cache -o --output"
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

_gts_pick()
{
    opts="-h --help --version -f --feature -F --format --no-cache -o --output"
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
    opts="-h --help --version -d --delimiter --empty --no-cache --no-header --no-key -n --name --no-location -o --output --source -t --separator"
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

_gts_repair()
{
    opts="-h --help --version -F --format --no-cache -o --output"
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
    opts="-h --help --version -F --format --no-cache -o --output"
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
    opts="-h --help --version -F --format --no-cache -o --output"
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
    opts="-h --help --version -e --exact -F --format -k --key --no-cache --no-complement -o --output -q --qualifier"
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
    opts="-h --help --version -F --format --no-cache -o --output -v --invert-match"
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

_gts_sort()
{
    opts="-h --help --version -F --format --no-cache -o --output -r --reverse"
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

_gts_split()
{
    opts="-h --help --version -F --format --no-cache -o --output"
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
    opts="-h --help --version -F --no-feature --no-cache -o --output -Q --no-qualifier"
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
    cmds="-h --help --version annotate cache clear complement define delete extract infix insert join length pick query repair reverse rotate search select sort split summary"
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
        cache)      _gts_cache ;;
        clear)      _gts_clear ;;
        complement) _gts_complement ;;
        define)     _gts_define ;;
        delete)     _gts_delete ;;
        extract)    _gts_extract ;;
        infix)      _gts_infix ;;
        insert)     _gts_insert ;;
        join)       _gts_join ;;
        length)     _gts_length ;;
        pick)       _gts_pick ;;
        query)      _gts_query ;;
        repair)     _gts_repair ;;
        reverse)    _gts_reverse ;;
        rotate)     _gts_rotate ;;
        search)     _gts_search ;;
        select)     _gts_select ;;
        sort)       _gts_sort ;;
        split)      _gts_split ;;
        summary)    _gts_summary ;;
        *) ;;
    esac
}

complete -F _gts gts