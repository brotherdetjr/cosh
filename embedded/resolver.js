(function() {
	var self;
	self = function(other) {
		join(self, other);
		return other;
	};
	self.link = newCmdLink.apply(null, arguments);
	return self;
});