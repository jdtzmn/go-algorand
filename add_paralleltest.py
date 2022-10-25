def open_log_file(log_file):
    """Opens the golangci log file from running `make lint > output.log`."""
    with open(log_file, "r") as f:
        lines = f.readlines()
    return lines


def is_paralleltest_line(line: str):
    indicator_substring = "(paralleltest)"
    return indicator_substring in line

def is_fixable(line: str):
    unfixable_substrings = ["has missing the call to method parallel"]
    return all(unfixable_substring not in line for unfixable_substring in unfixable_substrings)


def extract_file_path_and_line(line: str):
    """Given a paralleltest line, extract the file path and line as a tuple.
        This ignores the edge case where a `:` could be included in filepath on windows."""
    return tuple(line.split(":")[:2])


def parse_log_file(log_file):
    """Read the log file and parse it to find locations for paralleltest errors,
        augmented by what the new line number would be after fixing previous ones."""
    lines = open_log_file(log_file)

    # Preprocessed is a dictionary which maps file paths to a set of line numbers.
    # Line numbers correspond to the test's function declaration.
    preprocessed = {}

    for line in lines:
        if is_paralleltest_line(line) and is_fixable(line):
            file, line_number = extract_file_path_and_line(line)
            processed_lines = preprocessed.setdefault(file, set())
            processed_lines.add(int(line_number))

    return preprocessed


def fix_locations(file, line_numbers):
    """Fix location by adding `t.Parallel()` one line below where the error occured."""
    new_file_lines = []

    with open(file, "r") as f:
        # Use enumerate to minimize memory usage
        prev_line_tabs = 0
        for i, line in enumerate(f):
            if i + 1 in line_numbers: # Error line number
                prev_line_tabs = len(line) - len(line.lstrip('\t'))
            if i in line_numbers: # Fix line number
                tab_number = prev_line_tabs + 1
                t_parallel = tab_number * "\t" + "t.Parallel()" + "\n"
                new_file_lines.append(t_parallel)
            new_file_lines.append(line)

    with open(file, "w") as f:
        f.writelines(new_file_lines)


if __name__ == "__main__":
    log_file = "output.log"
    fixable_locations = parse_log_file(log_file)

    for file, line_numbers in fixable_locations.items():
        fix_locations(file, line_numbers)
