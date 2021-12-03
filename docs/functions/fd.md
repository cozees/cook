# File and Directory Functions

File and Directory functions provide several pre-define functionality create, delete or modified ones or more files and directories.

1. [rm](#rm)
2. [mkdir](#mkdir)
3. [rmdir](#rmdir)
4. [chmod](#chmod)
5. [chown](#chown)
6. [cp, copy](#cp-copy)
7. [mv, move](#mv-move)
8. [workin, chdir](#workin-chdir)
## @rm

Usage:
```cook
@rm [-p] PATH
```

Remove one or more files and directory in the hierarchy. If the given path is a file then only that that is remove       however it was a directory then it content including the directory itself will be remove.It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -r, --recursive | false | Remove all file or directory in hierarchy of the given directory. |

Example:

```cook
@rm -p dir1/dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @mkdir

Usage:
```cook
@mkdir [-p] [-m permission] PATH
```

Create one or multiple directories. The function utilize linux permission mode syntax to set permission. It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -p, --recursive | false | Create directories recursively if any directory in the given path is not exist.        By default, if permission mode is not given then a permission 740 is used. |
| --m, --mode |  | Set directory permission. The linux permission syntax is required in order provide the permission other than default permission 740. |

Example:

```cook
@mkdir -p -m 755 dir1/dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @rmdir

Usage:
```cook
@rmdir [-p] PATH
```

Remove one or multiply empty directory in the hierarchy. It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -p, --recursive | false | Remove all empty child directories include current directory as well. |

Example:

```cook
@rmdir -p dir1/dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @chmod

Usage:
```cook
@chown [-r] mode PATH [PATH ...]
```

Change permission mode of files or directories.It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -r, --recursive | false | Tell @chmode to change permission of all file or directory in the hierarchy. |

Example:

```cook
@chown -r u+x,g-w dir1 dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @chown

Usage:
```cook
@chown [-rn] [user][:group] PATH
```

Change owner and/or group of file or directory.It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -n, --guinum | false | Tell @chown that the given user and/or group id is a numeric id.                 By default, @chown treat the given user or group as a username or group name                which required lookup to find a numeric representation of user or group id. |
| -r, --recursive | false | Tell @chown to change owner of all file or directory in the hierarchy. |

Example:

```cook
@chown -r user1:group1 dir1/dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @cp, @copy

Usage:
```cook
@cp [-r] mode PATH [PATH ...] NEW_PATH
```

Copy one or more of files or directories. If the target is not exist the @cp will create like call @mkdir -p. It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |
| -r, --recursive | false | Copies the directory and the entire sub-tree to the target. To copy the content only add trailing /. |

Example:

```cook
@cp dir1 file.txt dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @mv, @move

Usage:
```cook
@mv TARGET_PATH DESTRINATION_PATH
```

Move a target file or directory to the destination path which must be an existed directory.It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@mv dir1 dir2/dir3
```
[back top](#file-and-directory-functions)

---

## @workin, @chdir

Usage:
```cook
@workin PATH
```

Change current working directory to the given directory.It fine to use linux file path syntax on any platform.

| Options/Flag | Default | Description |
| --- | --- | --- |

Example:

```cook
@workin dir1/dir2/dir3
```
[back top](#file-and-directory-functions)

---

