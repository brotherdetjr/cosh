(function () {
	var self;
	self = function (other) {
		return join(self, other);
	};
	self.link = newCmdLink.apply(null, arguments);
	return self;
});