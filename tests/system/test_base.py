from cryptocompbeat import BaseTest

import os


class Test(BaseTest):

    def test_base(self):
        """
        Basic test with exiting Cryptocompbeat normally
        """
        self.render_config_template(
            path=os.path.abspath(self.working_dir) + "/log/*"
        )

        cryptocompbeat_proc = self.start_beat()
        self.wait_until(lambda: self.log_contains("cryptocompbeat is running"))
        exit_code = cryptocompbeat_proc.kill_and_wait()
        assert exit_code == 0
