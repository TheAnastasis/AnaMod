package rpcs3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"anamod/internal/atomicfile"
	"anamod/internal/datapack"
	"go.yaml.in/yaml/v3"
)

func SetTicks(configYAML []byte, info datapack.RPCS3Info, ticks map[string]bool) ([]byte, error) {
	doc, root, err := parseMapping(configYAML)
	if err != nil {
		return nil, fmt.Errorf("patch_config.yml: %w", err)
	}
	for _, name := range slices.Sorted(maps.Keys(ticks)) {
		path := entryPath(info, name)
		if !ticks[name] {
			untick(root, path)
			continue
		}
		entry := root
		for _, key := range path {
			entry = ensureMapping(entry, key)
		}
		setScalar(entry, "Enabled", "true", "!!bool")
	}
	return encode(doc)
}

func entryPath(info datapack.RPCS3Info, patchName string) []string {
	return []string{info.PPUHash, patchName, info.Title, info.Serial, info.AppVersion}
}

func untick(node *yaml.Node, path []string) {
	child := mappingValue(node, path[0])
	if child == nil {
		return
	}
	if len(path) > 1 {
		if child.Kind != yaml.MappingNode {
			return
		}
		untick(child, path[1:])
		if len(child.Content) > 0 {
			return
		}
	}
	removeKey(node, path[0])
}

func IsTicked(configYAML []byte, info datapack.RPCS3Info, patchName string) bool {
	_, root, err := parseMapping(configYAML)
	if err != nil {
		return false
	}
	entry := root
	for _, key := range append(entryPath(info, patchName), "Enabled") {
		if entry == nil || entry.Kind != yaml.MappingNode {
			return false
		}
		entry = mappingValue(entry, key)
	}
	var enabled bool
	return entry != nil && entry.Decode(&enabled) == nil && enabled
}

func PatchNames(patchYAML []byte, ppuHash string) []string {
	_, root, err := parseMapping(patchYAML)
	if err != nil {
		return nil
	}
	patches := mappingValue(root, ppuHash)
	if patches == nil || patches.Kind != yaml.MappingNode {
		return nil
	}
	var names []string
	for i := 0; i+1 < len(patches.Content); i += 2 {
		names = append(names, patches.Content[i].Value)
	}
	return names
}

func HasPatch(patchYAML []byte, ppuHash, patchName string) bool {
	return slices.Contains(PatchNames(patchYAML, ppuHash), patchName)
}

func DuplicateNames(patchYAML []byte, ppuHash string, names []string) []string {
	var found []string
	for _, name := range PatchNames(patchYAML, ppuHash) {
		if slices.Contains(names, name) && !slices.Contains(found, name) {
			found = append(found, name)
		}
	}
	return found
}

func fpsOption(info datapack.RPCS3Info) *datapack.Option {
	for i := range info.Options {
		if info.Options[i].FallbackPatch != "" {
			return &info.Options[i]
		}
	}
	return nil
}

func ChooseTicks(info datapack.RPCS3Info, chosen []string, officialFPSExists bool) map[string]bool {
	ticks := map[string]bool{}
	for _, option := range info.Options {
		on := slices.Contains(chosen, option.ID)
		if option.FallbackPatch == "" {
			ticks[option.Patch] = on
			continue
		}
		ticks[option.Patch] = on && officialFPSExists
		ticks[option.FallbackPatch] = on && !officialFPSExists
	}
	return ticks
}

func perGamePatchPath(loc Location, info datapack.RPCS3Info) string {
	return filepath.Join(loc.PatchesDir(), info.Serial+"_patch.yml")
}

func Apply(loc Location, info datapack.RPCS3Info, patchFile []byte, chosen []string,
	log io.Writer) (officialFPSWasOn bool, warnings []string, err error) {
	configPath := loc.PatchConfigPath()
	config, err := readOptional(configPath)
	if err != nil {
		return false, nil, err
	}
	fps := fpsOption(info)
	officialExists := false
	if fps != nil {
		officialFPSWasOn = IsTicked(config, info, fps.Patch)
		list, err := readOptional(filepath.Join(loc.PatchesDir(), "patch.yml"))
		if err != nil {
			return false, nil, err
		}
		officialExists = HasPatch(list, info.PPUHash, fps.Patch)
		fmt.Fprintf(log, "RPCS3's own %q: in its patch list %v, ticked %v\n", fps.Patch,
			officialExists, officialFPSWasOn)
	}
	newConfig, err := SetTicks(config, info, ChooseTicks(info, chosen, officialExists))
	if err != nil {
		return false, nil, err
	}

	ours := PatchNames(patchFile, info.PPUHash)
	for _, other := range []string{"imported_patch.yml", "patch.yml"} {
		data, err := readOptional(filepath.Join(loc.PatchesDir(), other))
		if err != nil {
			return false, nil, err
		}
		for _, name := range DuplicateNames(data, info.PPUHash, ours) {
			warnings = append(warnings, fmt.Sprintf("RPCS3's %s also has a patch named %q, "+
				"(probably an older version??). Remove it in RPCS3's Patch Manager so only "+
				"AnaMod's current one is used.", other, name))
		}
	}
	for _, warning := range warnings {
		fmt.Fprintf(log, "Warning: %s\n", warning)
	}

	backup := configPath + ".anamod-backup"
	if config != nil && !exists(backup) {
		if err := os.WriteFile(backup, config, 0o644); err != nil {
			return false, nil, fmt.Errorf("could not back up %s: %w", configPath, err)
		}
		fmt.Fprintf(log, "Backed up %s\n", configPath)
	}
	perGame := perGamePatchPath(loc, info)
	oldPerGame, err := readOptional(perGame)
	if err != nil {
		return false, nil, err
	}
	if err := atomicfile.Write(perGame, patchFile); err != nil {
		return false, nil, fmt.Errorf("could not write %s: %w", perGame, err)
	}
	fmt.Fprintf(log, "Wrote %s\n", perGame)
	if err := atomicfile.Write(configPath, newConfig); err != nil {
		if oldPerGame != nil {
			atomicfile.Write(perGame, oldPerGame)
		} else {
			os.Remove(perGame)
		}
		return false, nil, fmt.Errorf("could not write %s: %w", configPath, err)
	}
	fmt.Fprintf(log, "Wrote %s (ticked: %v)\n", configPath, chosen)
	return officialFPSWasOn, warnings, nil
}

func Remove(loc Location, info datapack.RPCS3Info, officialFPSWasOn bool, log io.Writer) error {
	configPath := loc.PatchConfigPath()
	config, err := readOptional(configPath)
	if err != nil {
		return err
	}
	ticks := ChooseTicks(info, nil, false)
	if fps := fpsOption(info); fps != nil {
		ticks[fps.Patch] = officialFPSWasOn
	}
	newConfig, err := SetTicks(config, info, ticks)
	if err != nil {
		return err
	}
	if err := atomicfile.Write(configPath, newConfig); err != nil {
		return fmt.Errorf("could not write %s: %w", configPath, err)
	}
	fmt.Fprintf(log, "Wrote %s (AnaMod's patches unticked)\n", configPath)
	perGame := perGamePatchPath(loc, info)
	if err := os.Remove(perGame); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("could not delete %s: %w", perGame, err)
	}
	fmt.Fprintf(log, "Deleted %s\n", perGame)
	return nil
}

func parseMapping(data []byte) (*yaml.Node, *yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, nil, err
	}
	if doc.Kind == 0 || len(doc.Content) == 0 {
		root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}, root, nil
	}
	root := doc.Content[0]
	if root.Kind == yaml.ScalarNode && root.Tag == "!!null" {
		*root = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	}
	if root.Kind != yaml.MappingNode {
		return nil, nil, errors.New("the file is not a list of settings")
	}
	return &doc, root, nil
}

func encode(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func ensureMapping(mapping *yaml.Node, key string) *yaml.Node {
	if value := mappingValue(mapping, key); value != nil {
		if value.Kind != yaml.MappingNode {
			*value = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		}
		return value
	}
	value := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
	return value
}

func setScalar(mapping *yaml.Node, key, value, tag string) {
	if existing := mappingValue(mapping, key); existing != nil {
		*existing = yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value}
		return
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value})
}

func removeKey(mapping *yaml.Node, key string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
}

func readOptional(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", path, err)
	}
	return data, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
