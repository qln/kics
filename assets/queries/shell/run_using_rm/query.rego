package Cx

CxPolicy[result] {
	resource := input.document[i].command[name][idx]

	contains(resource.Value, "rm")

	result := {
		"documentId": input.document[i].id,
		"searchKey": sprintf("{{%s}}", [resource.Original]),
		"issueType": "IncorrectValue",
		"keyExpectedValue": "RUN instructions should not use the 'rm' program",
		"keyActualValue": "RUN instruction is invoking the 'rm' program",
	}
}
