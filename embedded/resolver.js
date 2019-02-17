(function() {
	var self;
	self = function(other) {
		join(self, other);
		return self;
	};
	self.link = newCmdLink.apply(null, arguments);
	return self;
});