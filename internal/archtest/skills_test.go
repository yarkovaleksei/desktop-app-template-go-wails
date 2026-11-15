package archtest

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// Скиллы агентов: каталоги в skillsDir, для Claude Code — относительные
// симлинки в claudeSkillsDir, список — таблица раздела «Скиллы» в AGENTS.md.
const (
	skillsDir       = ".agents/skills"
	claudeSkillsDir = ".claude/skills"
)

// skillRowRe — строка таблицы скиллов: имя в первой ячейке, ссылкой или в
// обратных кавычках.
var skillRowRe = regexp.MustCompile("^\\|\\s*(?:\\[([^\\]]+)\\]\\([^)]*\\)|`([^`]+)`)\\s*\\|")

// TestSkills требует, чтобы оба агента находили одни и те же скиллы:
//   - у каждого каталога .agents/skills/<имя> есть SKILL.md, во frontmatter
//     которого только name, равное имени каталога, и непустой description;
//   - .claude/skills/<имя> — относительный симлинк на ../../.agents/skills/<имя>,
//     лишних записей в .claude/skills нет;
//   - таблица раздела «Скиллы» в AGENTS.md перечисляет те же скиллы.
func TestSkills(t *testing.T) {
	root := moduleRoot(t)
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(skillsDir)))
	if err != nil {
		t.Fatal(err)
	}
	var skills []string
	for _, e := range entries {
		if !e.IsDir() {
			t.Errorf("%s/%s: в каталоге скиллов только каталоги", skillsDir, e.Name())
			continue
		}
		skills = append(skills, e.Name())
		checkSkillFile(t, root, e.Name())
		checkSkillLink(t, root, e.Name())
	}
	if len(skills) == 0 {
		t.Fatalf("%s: нет ни одного скилла", skillsDir)
	}

	links, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(claudeSkillsDir)))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range links {
		if !slices.Contains(skills, l.Name()) {
			t.Errorf("%s/%s: скилла %s/%s нет", claudeSkillsDir, l.Name(), skillsDir, l.Name())
		}
	}

	listed := agentsSkills(t, root)
	for _, s := range skills {
		if !slices.Contains(listed, s) {
			t.Errorf("AGENTS.md: скилла %s нет в таблице раздела «Скиллы»", s)
		}
	}
	for _, s := range listed {
		if !slices.Contains(skills, s) {
			t.Errorf("AGENTS.md: в таблице «Скиллы» есть %s, а каталога %s/%s нет", s, skillsDir, s)
		}
	}
}

// checkSkillFile проверяет frontmatter SKILL.md скилла name.
func checkSkillFile(t *testing.T, root, name string) {
	t.Helper()
	rel := skillsDir + "/" + name + "/SKILL.md"
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Errorf("%s: %v", rel, err)
		return
	}
	fields, ok := frontmatter(string(data))
	if !ok {
		t.Errorf("%s: файл начинается не с frontmatter между строками ---", rel)
		return
	}
	for key := range fields {
		if key != "name" && key != "description" {
			t.Errorf("%s: во frontmatter лишнее поле %q — только name и description", rel, key)
		}
	}
	if fields["name"] != name {
		t.Errorf("%s: name %q не совпадает с именем каталога %q", rel, fields["name"], name)
	}
	if strings.TrimSpace(fields["description"]) == "" {
		t.Errorf("%s: пустой description — по нему агент решает, когда применять скилл", rel)
	}
}

// frontmatter разбирает плоский frontmatter `ключ: значение` между строками
// `---` в начале файла; значение в двойных или одинарных кавычках
// раскавычивается.
func frontmatter(text string) (map[string]string, bool) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, false
	}
	fields := map[string]string{}
	for _, l := range lines[1:] {
		if l == "---" {
			return fields, true
		}
		key, value, ok := strings.Cut(l, ":")
		if !ok || strings.HasPrefix(l, " ") {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		fields[strings.TrimSpace(key)] = value
	}
	return nil, false
}

// checkSkillLink проверяет симлинк .claude/skills/<name>. Git на Windows без
// core.symlinks записывает симлинк обычным файлом с путём цели; там такой
// файл принимается, если путь верный.
func checkSkillLink(t *testing.T, root, name string) {
	t.Helper()
	rel := claudeSkillsDir + "/" + name
	want := "../../" + skillsDir + "/" + name
	file := filepath.Join(root, filepath.FromSlash(rel))
	info, err := os.Lstat(file)
	if errors.Is(err, fs.ErrNotExist) {
		t.Errorf("%s: нет симлинка на %s — без него Claude Code не видит скилл", rel, want)
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	var target string
	switch {
	case info.Mode()&fs.ModeSymlink != 0:
		if target, err = os.Readlink(file); err != nil {
			t.Fatal(err)
		}
	case runtime.GOOS == "windows" && info.Mode().IsRegular():
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		target = string(data)
	default:
		t.Errorf("%s: не симлинк", rel)
		return
	}
	if filepath.ToSlash(target) != want {
		t.Errorf("%s: ведёт на %q, а должен на %q (относительный путь)", rel, target, want)
		return
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		if st, err := os.Stat(file); err != nil || !st.IsDir() {
			t.Errorf("%s: симлинк не ведёт на каталог скилла", rel)
		}
	}
}

// agentsSkills возвращает имена скиллов из таблицы раздела «Скиллы» в AGENTS.md.
func agentsSkills(t *testing.T, root string) []string {
	t.Helper()
	lines, _ := readDocLines(t, filepath.Join(root, "AGENTS.md"))
	var names []string
	in := false
	for _, l := range lines {
		if m := headingRe.FindStringSubmatch(l); m != nil {
			in = m[1] == "##" && m[2] == "Скиллы"
			continue
		}
		if !in {
			continue
		}
		if m := skillRowRe.FindStringSubmatch(l); m != nil {
			names = append(names, m[1]+m[2])
		}
	}
	if len(names) == 0 {
		t.Fatal("AGENTS.md: в разделе «Скиллы» нет таблицы скиллов")
	}
	return names
}
