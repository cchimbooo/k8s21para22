package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func main() {
	if err := filepath.Walk("./", visit); err != nil {
		panic(err)
	}
	fmt.Println("done")

	time.Sleep(10)

}

func visit(path string, fi os.FileInfo, err error) error {

	if err != nil {
		return err
	}

	if !!fi.IsDir() {
		return nil //
	}

	matched, err := filepath.Match("deploy-staging.yaml", fi.Name())

	if err != nil {
		panic(err)
		return err
	}

	if matched {
		atualizardeploy(path)
	}
	return nil
}

func atualizardeploy(path string) {
	// setup reader
	f, err := os.Open(path)
	defer func() { _ = f.Close() }()
	if err != nil {
		panic(err)
	}

	d := yaml.NewDecoder(f)
	var fileContent []byte
	// Le o arquivo
	for {
		// create new spec here
		spec := yaml.MapSlice{}
		// pass a reference to spec reference
		errDecode := d.Decode(&spec)
		// check it was parsed
		if spec == nil {
			continue
		}
		// break the loop in case of EOF
		if errors.Is(errDecode, io.EOF) {
			break
		}

		// Converte para []bytes
		bYaml, err := yaml.Marshal(spec)
		if err != nil {
			panic(err.Error())
		}
		if validaSeConverteIngress(spec) {
			oldIngres, errCast := mapaParaStructIngress(spec)
			if errCast != nil {
				panic(errCast)
			}
			novoIngress := oldIngres.ConverterParaNovo()
			var e2 error
			bYaml, e2 = novoIngress.ToBYaml()
			if e2 != nil {
				panic(e2)
			}

		}
		fileContent = append(fileContent, concatenate(bYaml)...)

		if err != nil {
			panic(err)
		}
	}
	arquivoxuxado := fincaString(removeNull(fileContent))
	fmt.Println(string(arquivoxuxado))
	writeToFile(arquivoxuxado, path)
	check(fileContent)
}

func concatenate(bYaml []byte) []byte {
	v := "\n---\n"
	bYaml = append(bYaml, []byte(v)...)
	return bYaml
}

func writeToFile(bYaml []byte, path string) {

	if err := ioutil.WriteFile(path, bYaml, 0); err != nil {
		panic(err)
	}

}

/*func xuxaConfigMap(b yaml.MapSlice) yaml.MapSlice {
    alterar := false
    for _, v := range b {
        switch v.Key.(type) {
        case string:
            if v.Key.(string) == "kind" {
                switch v.Value.(type) {
                case string:
                    if v.Value.(string) == "ConfigMap" {
                        alterar = true
                    }
                }
            }
        }
    }
    if alterar {
        for k1, v := range b {
            switch v.Key.(type) {
            case string:
                if v.Key.(string) == "data" {
                    switch v.Value.(type) {
                    case yaml.MapSlice:
                        ms := v.Value.(yaml.MapSlice)
                        for k2, item := range ms {
                            switch item.Value.(type) {
                            case string:

                                s := "\"" + item.Value.(string) + "\""
                                var i interface{}
                                i = s
                                b[k1].Value.(yaml.MapSlice)[k2].Value = i
                                fmt.Println(b[k1].Value.(yaml.MapSlice)[k2].Value)
                            }
                        }
                    }
                }
            }
        }
    }
    return b
}*/

func validaSeConverteIngress(b yaml.MapSlice) bool {

	apiVersion := false
	kind := false
	for _, v := range b {
		switch v.Key.(type) {
		case string:
			if v.Key.(string) == "apiVersion" {
				switch v.Value.(type) {
				case string:
					if v.Value.(string) == "extensions/v1beta1" {
						apiVersion = true
					}
				}
			}
			if v.Key.(string) == "kind" {
				switch v.Value.(type) {
				case string:
					if v.Value.(string) == "Ingress" {
						kind = true
					}
				}
			}
		}
	}
	return apiVersion && kind
}

func mapaParaStructIngress(b yaml.MapSlice) (IngressStruct, error) {
	i := IngressStruct{}

	m, e := yaml.Marshal(b)
	if e != nil {
		return IngressStruct{}, e
	}

	if err := yaml.Unmarshal(m, &i); err != nil {
		if err != nil {
			panic(err.Error())
		}
		return i, err
	}
	return i, nil

}

func check(b []byte) {
	s := []string{
		"admissionregistration.k8s.io/v1beta1",
		"apiextensions.k8s.io/v1beta1",
		"apiregistration.k8s.io/v1beta1",
		"authentication.k8s.io/v1beta1",
		"authorization.k8s.io/v1beta1",
		"certificates.k8s.io/v1beta1",
		"coordination.k8s.io/v1beta1",
		"extensions/v1beta1",
		"networking.k8s.io/v1beta1",
		"networking.k8s.io/v1beta1",
		"rbac.authorization.k8s.io/v1beta1",
		"scheduling.k8s.io/v1beta1",
		"storage.k8s.io/v1beta1",
	}
	fuu := []string{}
	for _, v := range s {
		if bytes.Contains(b, []byte(v)) {
			fuu = append(fuu, v)
		}
	}
	if len(fuu) > 0 {
		fmt.Println("Existem dados errados \\o/, contate o Felipe.")
		for _, v := range fuu {
			fmt.Println(v)
		}
		time.Sleep(600 * time.Second)
	}
}

func fincaString(b []byte) []byte {
	exp := "(\\$\\{.*\\})"
	r, err := regexp.Compile(exp)
	if err != nil {
		panic(err)
	}
	k := r.FindAll(b, -1)
	spew.Dump(k)
	nw := r.ReplaceAllFunc(
		b, func(i []byte) []byte {
			nb := make([]byte, len(i)+2)
			nb[0] = '"'
			for k, v := range i {
				nb[k+1] = v
			}
			nb[len(i)+1] = '"'
			return nb
		},
	)
	return nw
}

func removeNull(b []byte) []byte {
	r, err := regexp.Compile(`(null)`)
	if err != nil {
		panic(err)
	}
	return r.ReplaceAll(b, []byte{})
}
