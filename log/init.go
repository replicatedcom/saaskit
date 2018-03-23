package log

func Init(logOpts *LogOptions, mailLogOpts *MailLogOptions, slackLogOpts *SlackLogOptions) {
	InitLog(logOpts)
	InitMail(mailLogOpts)
	InitSlack(slackLogOpts)
}
