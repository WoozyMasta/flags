$__goFlagsCommand = 'examples'

Register-ArgumentCompleter -Native -CommandName $__goFlagsCommand -ScriptBlock {
	param($wordToComplete, $commandAst, $cursorPosition)

	$elements = @($commandAst.CommandElements | ForEach-Object { $_.Extent.Text })
	if ($elements.Count -eq 0) {
		return
	}

	$exe = $elements[0]
	$args = @()
	if ($elements.Count -gt 1) {
		$args = $elements[1..($elements.Count - 1)]
	}

	$prev = $env:GO_FLAGS_COMPLETION
	$env:GO_FLAGS_COMPLETION = '1'
	try {
		$items = & $exe @args
	} finally {
		if ($null -ne $prev) {
			$env:GO_FLAGS_COMPLETION = $prev
		} else {
			Remove-Item Env:GO_FLAGS_COMPLETION -ErrorAction SilentlyContinue
		}
	}

	foreach ($item in $items) {
		[System.Management.Automation.CompletionResult]::new(
			$item,
			$item,
			'ParameterValue',
			$item
		)
	}
}
