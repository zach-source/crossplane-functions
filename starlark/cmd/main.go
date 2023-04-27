package main

import (
	"fmt"
	"io"
	"os"

	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	cuelangv1alpha1 "github.com/zach-source/crossplane-functions/starlark/api/v1alpha1"
)

var (
	Colors = []string{"red", "green", "blue", "yellow", "orange", "purple", "black", "white"}
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v", err)
		os.Exit(1)
	}
	obj := &v1alpha1.FunctionIO{}
	if err := yaml.Unmarshal(b, obj); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal stdin: %v", err)
		os.Exit(1)
	}

	config := cuelangv1alpha1.Config{}

	if err := json.Unmarshal(obj.Config.Raw, &config); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal stdin: %v", err)
		os.Exit(1)
	}

	c := cuecontext.New()
	v := c.CompileString(config.Spec.Template, cue.Filename("schema.cue"))

	// check for errors during compiling
	if v.Err() != nil {
		msg := errors.Details(v.Err(), nil)
		fmt.Printf("Compile Error:\n%s\n", msg)
	}

	// To get all errors, we need to validate
	err = v.Validate()
	if err != nil {
		msg := errors.Details(err, nil)
		fmt.Printf("Validate Error:\n%s\n", msg)
	}

	observedComposite := map[string]any{}
	yaml.Unmarshal(obj.Observed.Composite.Resource.Raw, &observedComposite)
	v = v.FillPath(cue.ParsePath("observed.composite"), observedComposite)
	str, err := v.Lookup("observed").MarshalJSON()
	if err != nil {
		msg := errors.Details(err, nil)
		fmt.Printf("Validate Error:\n%s\n", msg)
	}
	fmt.Printf("%s\n", str)

	str, err = v.Lookup("desired").MarshalJSON()
	if err != nil {
		msg := errors.Details(err, nil)
		fmt.Printf("Validate Error:\n%s\n", msg)
	}
	fmt.Printf("%s\n", str)

	err = json.Unmarshal(str, &obj.Desired)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal resulting functionio: %v", err)
		os.Exit(1)
	}

	result, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal resulting functionio: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(result))
	// robotGroup := composite.New()
	// if err := yaml.Unmarshal(obj.Observed.Composite.Resource.Raw, &robotGroup.Unstructured); err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to unmarshal observed composite: %v", err)
	// 	os.Exit(1)
	// }
	// count, err := fieldpath.Pave(robotGroup.Object).GetInteger("spec.count")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to get the count from observed composite: %v", err)
	// 	os.Exit(1)
	// }
	// var robots []v1alpha1.DesiredResource
	// for _, r := range obj.Observed.Resources {
	// 	robots = append(robots, v1alpha1.DesiredResource{
	// 		Name:     r.Name,
	// 		Resource: r.Resource,
	// 	})
	// }
	// add := int(count) - len(robots)
	// for i := 0; i < add; i++ {
	// 	suf, err := generateSuffix()
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "failed to generate random suffix for name: %v", err)
	// 		os.Exit(1)
	// 	}
	// 	r := &dummyv1alpha1.Robot{
	// 		Spec: dummyv1alpha1.RobotSpec{
	// 			ForProvider: dummyv1alpha1.RobotParameters{
	// 				Color: Colors[rand.Intn(len(Colors))],
	// 			},
	// 		},
	// 	}
	// 	r.SetName(robotGroup.GetName() + "-" + suf)
	// 	r.SetGroupVersionKind(dummyv1alpha1.RobotGroupVersionKind)
	// 	// NOTE: We need to use a JSON marshaller here because runtiem.RawExtension
	// 	// type expects a JSON blob.
	// 	raw, err := json.Marshal(r)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "failed to marshal resource: %v", err)
	// 		os.Exit(1)
	// 	}
	// 	robots = append(robots, v1alpha1.DesiredResource{
	// 		Name: "robot-" + suf,
	// 		Resource: runtime.RawExtension{
	// 			Raw: raw,
	// 		},
	// 	})
	// }
	// obj.Desired.Resources = robots
}

func generateSuffix() (string, error) {
	return password.Settings{
		CharacterSet: "abcdefghijklmnopqrstuvwxyz0123456789",
		Length:       5,
	}.Generate()
}
