package Cx

CxPolicy[result] {
	resource := input.document[i].command[name][idx]

	contains(resource.Value, "passwd")

	result := {
		"documentId": input.document[i].id,
		"searchKey": sprintf("{{%s}}", [resource.Original]),
		"issueType": "IncorrectValue",
		"keyExpectedValue": "RUN instructions should not use the 'plain password'",
		"keyActualValue": "RUN instruction is invoking the 'plain password'",
	}
}
