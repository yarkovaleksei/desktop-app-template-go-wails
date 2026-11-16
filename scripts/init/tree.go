package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// entry — отслеживаемый файл дерева: обычный файл или симлинк.
type entry struct {
	path string      // путь от корня через «/»
	from string      // исходный путь; пусто — файл создан инициализацией
	data []byte      // содержимое файла или цель симлинка
	link bool        // симлинк
	perm fs.FileMode // права обычного файла
}

// text — можно ли менять файл как текст: двоичные файлы содержат нулевой байт.
func (e *entry) text() bool {
	head := e.data
	if len(head) > 8000 {
		head = head[:8000]
	}
	return !bytes.Contains(head, []byte{0})
}

// git запускает git в каталоге root и возвращает его вывод.
func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// readTree читает отслеживаемые git файлы. Симлинк на ОС без симлинков
// (Windows с core.symlinks=false) git кладёт файлом с целью внутри: такой
// файл читается как текстовый и меняется по тем же правилам.
func readTree(root string) ([]*entry, error) {
	out, err := git(root, "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	var tree []*entry
	for _, p := range strings.Split(strings.TrimRight(out, "\x00"), "\x00") {
		if p == "" {
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(p))
		info, err := os.Lstat(full)
		if err != nil {
			return nil, err
		}
		e := &entry{path: p, from: p}
		if info.Mode()&fs.ModeSymlink != 0 {
			target, err := os.Readlink(full)
			if err != nil {
				return nil, err
			}
			e.link, e.data = true, []byte(filepath.ToSlash(target))
		} else {
			if e.data, err = os.ReadFile(full); err != nil {
				return nil, err
			}
			e.perm = info.Mode().Perm()
		}
		tree = append(tree, e)
	}
	return tree, nil
}

// change — одно изменение дерева для списка в выводе.
type change struct {
	kind string // изменить, переименовать, создать, удалить
	path string
	to   string // новый путь для переименования
}

func (c change) String() string {
	if c.to != "" {
		return fmt.Sprintf("  %-14s %s → %s", c.kind, c.path, c.to)
	}
	return fmt.Sprintf("  %-14s %s", c.kind, c.path)
}

// diff сравнивает исходное и новое дерево.
func diff(before, after []*entry) []change {
	old := map[string]*entry{}
	for _, e := range before {
		old[e.path] = e
	}
	kept := map[string]bool{}
	var changes []change
	for _, e := range after {
		if e.from == "" {
			changes = append(changes, change{kind: "создать", path: e.path})
			continue
		}
		kept[e.from] = true
		o := old[e.from]
		same := bytes.Equal(o.data, e.data) && o.link == e.link
		switch {
		case e.path != e.from && same:
			changes = append(changes, change{kind: "переименовать", path: e.from, to: e.path})
		case e.path != e.from:
			changes = append(changes, change{kind: "переименовать", path: e.from, to: e.path + " (и изменить)"})
		case !same:
			changes = append(changes, change{kind: "изменить", path: e.path})
		}
	}
	for _, e := range before {
		if !kept[e.path] {
			changes = append(changes, change{kind: "удалить", path: e.path})
		}
	}
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].path < changes[j].path })
	return changes
}

// apply записывает новое дерево на диск: новые и изменённые файлы, затем
// удаление исходных путей, которых в новом дереве нет, и опустевших каталогов.
func apply(root string, before, after []*entry) error {
	old := map[string]*entry{}
	for _, e := range before {
		old[e.path] = e
	}
	newPaths := map[string]bool{}
	for _, e := range after {
		newPaths[e.path] = true
	}
	for _, e := range after {
		if o := old[e.from]; e.from != "" && e.path == e.from && bytes.Equal(o.data, e.data) && o.link == e.link {
			continue
		}
		full := filepath.Join(root, filepath.FromSlash(e.path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if e.link {
			if err := os.Remove(full); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			if err := os.Symlink(filepath.FromSlash(string(e.data)), full); err != nil {
				return fmt.Errorf("симлинк %s: %w", e.path, err)
			}
			continue
		}
		perm := e.perm
		if perm == 0 {
			perm = 0o644
		}
		if err := os.WriteFile(full, e.data, perm); err != nil {
			return err
		}
	}
	var dirs []string
	for _, e := range before {
		if newPaths[e.path] {
			continue
		}
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(e.path))); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		dirs = append(dirs, path.Dir(e.path))
	}
	for _, d := range dirs {
		for ; d != "." && d != "/"; d = path.Dir(d) {
			if os.Remove(filepath.Join(root, filepath.FromSlash(d))) != nil {
				break // каталог не пуст или уже удалён
			}
		}
	}
	return nil
}
