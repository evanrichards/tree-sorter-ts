package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConstructorParameterSorting(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		changed bool
	}{
		{
			name: "basic_constructor_sorting",
			input: `class UserService {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly userRepository: UserRepository,
		private readonly logger: Logger,
		private readonly cache: CacheService,
		private readonly eventBus: EventBus,
	) {}
}`,
			want: `class UserService {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly cache: CacheService,
		private readonly eventBus: EventBus,
		private readonly logger: Logger,
		private readonly userRepository: UserRepository,
	) {}
}`,
			changed: true,
		},
		{
			name: "already_sorted_constructor",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly aService: AService,
		private readonly bService: BService,
		private readonly cService: CService,
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly aService: AService,
		private readonly bService: BService,
		private readonly cService: CService,
	) {}
}`,
			changed: false,
		},
		{
			name: "mixed_modifiers",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly zService: ZService,
		public aService: AService,
		protected readonly mService: MService,
		private bService: BService,
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		public aService: AService,
		private bService: BService,
		protected readonly mService: MService,
		private readonly zService: ZService,
	) {}
}`,
			changed: true,
		},
		{
			name: "with_comments",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		// Z service for handling Z operations
		private readonly zService: ZService,
		// A service for handling A operations
		private readonly aService: AService,
		private readonly mService: MService, // Middle service
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		// A service for handling A operations
		private readonly aService: AService,
		private readonly mService: MService, // Middle service
		// Z service for handling Z operations
		private readonly zService: ZService,
	) {}
}`,
			changed: true,
		},
		{
			name: "with_deprecated",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted deprecated-at-end **/
		private readonly newService: NewService,
		/** @deprecated Use newService instead */
		private readonly oldService: OldService,
		private readonly activeService: ActiveService,
		private readonly betaService: BetaService, // @deprecated
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted deprecated-at-end **/
		private readonly activeService: ActiveService,
		private readonly newService: NewService,
		private readonly betaService: BetaService, // @deprecated
		/** @deprecated Use newService instead */
		private readonly oldService: OldService,
	) {}
}`,
			changed: true,
		},
		{
			name: "with_new_line",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted with-new-line **/
		private readonly zService: ZService,
		private readonly aService: AService,
		private readonly mService: MService,
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted with-new-line **/
		private readonly aService: AService,

		private readonly mService: MService,

		private readonly zService: ZService,
	) {}
}`,
			changed: true,
		},
		{
			name: "optional_parameters",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly requiredZ: ZService,
		private readonly requiredA: AService,
		private readonly optionalB?: BService,
		private readonly optionalD?: DService,
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly optionalB?: BService,
		private readonly optionalD?: DService,
		private readonly requiredA: AService,
		private readonly requiredZ: ZService,
	) {}
}`,
			changed: true,
		},
		{
			name: "no_modifiers",
			input: `function createService(
	/** tree-sorter-ts: keep-sorted **/
	zParam: string,
	aParam: number,
	mParam: boolean,
) {}`,
			want: `function createService(
	/** tree-sorter-ts: keep-sorted **/
	aParam: number,
	mParam: boolean,
	zParam: string,
) {}`,
			changed: true,
		},
		{
			name: "single_parameter",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly service: Service,
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly service: Service,
	) {}
}`,
			changed: false,
		},
		{
			name: "no_magic_comment",
			input: `class Service {
	constructor(
		private readonly zService: ZService,
		private readonly aService: AService,
	) {}
}`,
			want: `class Service {
	constructor(
		private readonly zService: ZService,
		private readonly aService: AService,
	) {}
}`,
			changed: false,
		},
		{
			name: "trailing_comma_preserved",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly zService: ZService,
		private readonly aService: AService
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly aService: AService,
		private readonly zService: ZService
	) {}
}`,
			changed: true,
		},
		{
			name: "multiline_comment",
			input: `class Service {
	constructor(
		/**
		 * tree-sorter-ts: keep-sorted
		 *   with-new-line
		 *   deprecated-at-end
		 */
		private readonly active: ActiveService,
		/** @deprecated */
		private readonly old: OldService,
		private readonly beta: BetaService,
	) {}
}`,
			want: `class Service {
	constructor(
		/**
		 * tree-sorter-ts: keep-sorted
		 *   with-new-line
		 *   deprecated-at-end
		 */
		private readonly active: ActiveService,

		private readonly beta: BetaService,

		/** @deprecated */
		private readonly old: OldService,
	) {}
}`,
			changed: true,
		},
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "constructor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.name+".ts")
			err := os.WriteFile(testFile, []byte(tt.input), 0o644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Process the file
			result, err := ProcessFileAST(testFile, Config{Write: true})
			if err != nil {
				t.Fatalf("ProcessFileAST failed: %v", err)
			}

			if result.Changed != tt.changed {
				t.Errorf("Changed = %v, want %v", result.Changed, tt.changed)
			}

			// Read the processed content
			got, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// Normalize whitespace for comparison
			gotNorm := strings.TrimSpace(string(got))
			wantNorm := strings.TrimSpace(tt.want)

			if gotNorm != wantNorm {
				t.Errorf("Content mismatch:\ngot:\n%s\n\nwant:\n%s", string(got), tt.want)
			}
		})
	}
}

func TestConstructorEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		changed bool
	}{
		{
			name: "arrow_function_parameters",
			input: `const handler = (
	/** tree-sorter-ts: keep-sorted **/
	zParam: string,
	aParam: number,
	mParam: boolean,
) => {}`,
			want: `const handler = (
	/** tree-sorter-ts: keep-sorted **/
	aParam: number,
	mParam: boolean,
	zParam: string,
) => {}`,
			changed: true,
		},
		{
			name: "method_parameters",
			input: `class Service {
	process(
		/** tree-sorter-ts: keep-sorted **/
		zParam: string,
		aParam: number,
		mParam: boolean,
	) {}
}`,
			want: `class Service {
	process(
		/** tree-sorter-ts: keep-sorted **/
		aParam: number,
		mParam: boolean,
		zParam: string,
	) {}
}`,
			changed: true,
		},
		{
			name: "interface_method",
			input: `interface IService {
	process(
		/** tree-sorter-ts: keep-sorted **/
		zParam: string,
		aParam: number,
		mParam: boolean,
	): void;
}`,
			want: `interface IService {
	process(
		/** tree-sorter-ts: keep-sorted **/
		aParam: number,
		mParam: boolean,
		zParam: string,
	): void;
}`,
			changed: true,
		},
		{
			name: "complex_types",
			input: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly zService: Promise<ZService>,
		private readonly aService: Array<AService>,
		private readonly mService: Map<string, MService>,
	) {}
}`,
			want: `class Service {
	constructor(
		/** tree-sorter-ts: keep-sorted **/
		private readonly aService: Array<AService>,
		private readonly mService: Map<string, MService>,
		private readonly zService: Promise<ZService>,
	) {}
}`,
			changed: true,
		},
		{
			name: "destructured_parameters",
			input: `function process(
	/** tree-sorter-ts: keep-sorted **/
	{ zProp }: { zProp: string },
	{ aProp }: { aProp: number },
	{ mProp }: { mProp: boolean },
) {}`,
			want: `function process(
	/** tree-sorter-ts: keep-sorted **/
	{ aProp }: { aProp: number },
	{ mProp }: { mProp: boolean },
	{ zProp }: { zProp: string },
) {}`,
			changed: true,
		},
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "constructor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.name+".ts")
			err := os.WriteFile(testFile, []byte(tt.input), 0o644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Process the file
			result, err := ProcessFileAST(testFile, Config{Write: true})
			if err != nil {
				t.Fatalf("ProcessFileAST failed: %v", err)
			}

			if result.Changed != tt.changed {
				t.Errorf("Changed = %v, want %v", result.Changed, tt.changed)
			}

			// Read the processed content
			got, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			// Normalize whitespace for comparison
			gotNorm := strings.TrimSpace(string(got))
			wantNorm := strings.TrimSpace(tt.want)

			if gotNorm != wantNorm {
				t.Errorf("Content mismatch:\ngot:\n%s\n\nwant:\n%s", string(got), tt.want)
			}
		})
	}
}